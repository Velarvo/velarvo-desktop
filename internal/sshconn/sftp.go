package sshconn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/Velarvo/velarvo-desktop/internal/files"
	"github.com/pkg/sftp"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/crypto/ssh"
)

const (
	EventTransferProgress = "sftp:transfer:progress"

	progressInterval = 120 * time.Millisecond
)

type TransferProgress struct {
	TransferID string `json:"transferId"`
	Done       int64  `json:"done"`
	Total      int64  `json:"total"`
}

func init() {
	application.RegisterEvent[TransferProgress](EventTransferProgress)
}

type sftpClient struct {
	ssh  *ssh.Client
	sftp *sftp.Client
}

type SFTP struct {
	clients *Clients

	mu     sync.Mutex
	opened map[string]sftpClient
}

func NewSFTP(clients *Clients) *SFTP {
	return &SFTP{clients: clients, opened: make(map[string]sftpClient)}
}

func (s *SFTP) client(connectionID string) (*sftp.Client, error) {
	sshClient, err := s.clients.get(connectionID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.opened[connectionID]; ok {
		if c.ssh == sshClient {
			return c.sftp, nil
		}
		_ = c.sftp.Close()
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("open sftp: %w", err)
	}
	s.opened[connectionID] = sftpClient{ssh: sshClient, sftp: client}
	return client, nil
}

func (s *SFTP) Home(connectionID string) (string, error) {
	client, err := s.client(connectionID)
	if err != nil {
		return "", err
	}
	if wd, err := client.Getwd(); err == nil && wd != "" {
		return wd, nil
	}
	return "/", nil
}

func (s *SFTP) List(connectionID, dir string) ([]files.Entry, error) {
	client, err := s.client(connectionID)
	if err != nil {
		return nil, err
	}

	infos, err := client.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	entries := make([]files.Entry, 0, len(infos))
	for _, info := range infos {
		entry := files.Entry{
			Name:     info.Name(),
			Path:     path.Join(dir, info.Name()),
			Type:     files.TypeFile,
			Size:     info.Size(),
			Modified: info.ModTime().UnixMilli(),
		}
		isDir := info.IsDir()
		if info.Mode()&os.ModeSymlink != 0 {
			if target, err := client.Stat(entry.Path); err == nil {
				isDir = target.IsDir()
			}
		}
		if isDir {
			entry.Type, entry.Size = files.TypeDirectory, 0
		}
		entries = append(entries, entry)
	}
	files.SortEntries(entries)
	return entries, nil
}

func (s *SFTP) Mkdir(connectionID, dir string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	if _, err := client.Lstat(dir); err == nil {
		return files.ErrTargetExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return client.Mkdir(dir)
}

func (s *SFTP) CreateFile(connectionID, target string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	if _, err := client.Lstat(target); err == nil {
		return files.ErrTargetExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := client.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	return file.Close()
}

func (s *SFTP) Remove(connectionID, target string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	return client.RemoveAll(target)
}

func (s *SFTP) Rename(connectionID, from, to string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	return client.Rename(from, to)
}

func (s *SFTP) Copy(ctx context.Context, connectionID, from, to string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	src, info, err := openFile(client, from)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := client.Create(to)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if err := copyFile(ctx, dst, src, "", info.Size()); err != nil {
		_ = client.Remove(to)
		return err
	}
	_ = client.Chmod(to, info.Mode().Perm())
	return nil
}

func (s *SFTP) Preview(connectionID, remotePath string) (files.Preview, error) {
	client, err := s.client(connectionID)
	if err != nil {
		return files.Preview{}, err
	}
	file, info, err := openFile(client, remotePath)
	if err != nil {
		return files.Preview{}, err
	}
	defer func() { _ = file.Close() }()

	return files.BuildPreview(path.Base(remotePath), remotePath, info.Size(), file)
}

func (s *SFTP) Download(ctx context.Context, connectionID, remotePath, localPath, transferID string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}
	src, info, err := openFile(client, remotePath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	localPath = filepath.Clean(localPath)
	if err := os.MkdirAll(filepath.Dir(localPath), 0o750); err != nil {
		return err
	}
	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if err := copyFile(ctx, dst, src, transferID, info.Size()); err != nil {
		_ = dst.Close()
		_ = os.Remove(localPath)
		return err
	}
	return dst.Sync()
}

func (s *SFTP) Upload(ctx context.Context, connectionID, localPath, remotePath, transferID string) error {
	client, err := s.client(connectionID)
	if err != nil {
		return err
	}

	localPath = filepath.Clean(localPath)
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return files.ErrNotAFile
	}
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	if parent := path.Dir(remotePath); parent != "." && parent != "/" {
		_ = client.MkdirAll(parent)
	}
	dst, err := client.Create(remotePath)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if err := copyFile(ctx, dst, src, transferID, info.Size()); err != nil {
		_ = dst.Close()
		_ = client.Remove(remotePath)
		return err
	}
	_ = client.Chmod(remotePath, info.Mode().Perm())
	return nil
}

func openFile(client *sftp.Client, remotePath string) (*sftp.File, os.FileInfo, error) {
	info, err := client.Stat(remotePath)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, nil, files.ErrNotAFile
	}
	file, err := client.Open(remotePath)
	if err != nil {
		return nil, nil, err
	}
	return file, info, nil
}

func copyFile(ctx context.Context, dst io.Writer, src io.Reader, transferID string, total int64) error {
	w := &progressWriter{ctx: ctx, w: dst, id: transferID, total: total}
	if _, err := io.Copy(w, src); err != nil {
		return err
	}
	w.report()
	return nil
}

type progressWriter struct {
	ctx      context.Context
	w        io.Writer
	id       string
	done     int64
	total    int64
	reported time.Time
}

func (p *progressWriter) Write(b []byte) (int, error) {
	if err := p.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := p.w.Write(b)
	p.done += int64(n)
	if time.Since(p.reported) >= progressInterval {
		p.report()
	}
	return n, err
}

func (p *progressWriter) report() {
	p.reported = time.Now()
	if p.id != "" {
		emit(EventTransferProgress, TransferProgress{TransferID: p.id, Done: p.done, Total: p.total})
	}
}
