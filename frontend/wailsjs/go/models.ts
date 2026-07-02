export namespace app {

	export class ExplorerNode {
	    id: string;
	    name: string;
	    path: string;
	    type: string;
	    size?: number;
	    modified: number;

	    static createFrom(source: any = {}) {
	        return new ExplorerNode(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	    }
	}

}

export namespace projects {

	export class CreateProjectRequest {
	    name: string;
	    color?: string;

	    static createFrom(source: any = {}) {
	        return new CreateProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	}
	export class ProjectDTO {
	    id: string;
	    name: string;
	    color: string;
	    sortOrder: number;
	    hasIcon: boolean;
	    createdAt: number;
	    updatedAt: number;
	    revision: string;

	    static createFrom(source: any = {}) {
	        return new ProjectDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.sortOrder = source["sortOrder"];
	        this.hasIcon = source["hasIcon"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.revision = source["revision"];
	    }
	}
	export class ProjectIconDTO {
	    projectId: string;
	    status: string;
	    syncState: string;
	    mime?: string;
	    dataBase64?: string;
	    updatedAt?: number;

	    static createFrom(source: any = {}) {
	        return new ProjectIconDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.status = source["status"];
	        this.syncState = source["syncState"];
	        this.mime = source["mime"];
	        this.dataBase64 = source["dataBase64"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SetProjectIconRequest {
	    projectId: string;
	    mime: string;
	    data: number[];

	    static createFrom(source: any = {}) {
	        return new SetProjectIconRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.mime = source["mime"];
	        this.data = source["data"];
	    }
	}
	export class UpdateProjectRequest {
	    id: string;
	    name: string;
	    color?: string;

	    static createFrom(source: any = {}) {
	        return new UpdateProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	}

}

export namespace settings {

	export class Settings {
	    language: string;

	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	    }
	}

}

export namespace sshconn {

	export class ConnectionDTO {
	    id: string;
	    workspaceId: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    hasPassword: boolean;
	    os: string;
	    connected: boolean;
	    sortOrder: number;
	    lastUsedAt?: number;
	    createdAt: number;
	    updatedAt: number;
	    revision: string;

	    static createFrom(source: any = {}) {
	        return new ConnectionDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.hasPassword = source["hasPassword"];
	        this.os = source["os"];
	        this.connected = source["connected"];
	        this.sortOrder = source["sortOrder"];
	        this.lastUsedAt = source["lastUsedAt"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.revision = source["revision"];
	    }
	}
	export class ConnectionState {
	    id: string;
	    connected: boolean;
	    os: string;

	    static createFrom(source: any = {}) {
	        return new ConnectionState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.connected = source["connected"];
	        this.os = source["os"];
	    }
	}
	export class CreateConnectionRequest {
	    workspaceId: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;

	    static createFrom(source: any = {}) {
	        return new CreateConnectionRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceId = source["workspaceId"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class UpdateConnectionRequest {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    clearPassword: boolean;

	    static createFrom(source: any = {}) {
	        return new UpdateConnectionRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.clearPassword = source["clearPassword"];
	    }
	}

}

export namespace types {

	export class APIResponse___github_com_Velarvo_velarvo_desktop_internal_app_ExplorerNode_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: app.ExplorerNode[];
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse___github_com_Velarvo_velarvo_desktop_internal_app_ExplorerNode_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], app.ExplorerNode);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse___github_com_Velarvo_velarvo_desktop_internal_projects_ProjectDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: projects.ProjectDTO[];
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse___github_com_Velarvo_velarvo_desktop_internal_projects_ProjectDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], projects.ProjectDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse___github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: sshconn.ConnectionDTO[];
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse___github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], sshconn.ConnectionDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse___github_com_Velarvo_velarvo_desktop_internal_workspaces_WorkspaceDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: workspaces.WorkspaceDTO[];
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse___github_com_Velarvo_velarvo_desktop_internal_workspaces_WorkspaceDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], workspaces.WorkspaceDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse___string_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: string[];
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse___string_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = source["data"];
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_projects_ProjectDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: projects.ProjectDTO;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_projects_ProjectDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], projects.ProjectDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_projects_ProjectIconDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: projects.ProjectIconDTO;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_projects_ProjectIconDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], projects.ProjectIconDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_settings_Settings_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: settings.Settings;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_settings_Settings_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], settings.Settings);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: sshconn.ConnectionDTO;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], sshconn.ConnectionDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionState_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: sshconn.ConnectionState;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_sshconn_ConnectionState_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], sshconn.ConnectionState);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_vault_VaultState_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: vault.VaultState;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_vault_VaultState_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], vault.VaultState);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_github_com_Velarvo_velarvo_desktop_internal_workspaces_WorkspaceDTO_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: workspaces.WorkspaceDTO;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_github_com_Velarvo_velarvo_desktop_internal_workspaces_WorkspaceDTO_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = this.convertValues(source["data"], workspaces.WorkspaceDTO);
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIResponse_map_string_string_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: Record<string, string>;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_map_string_string_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = source["data"];
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }
	}
	export class APIResponse_string_ {
	    success: boolean;
	    code: string;
	    message: string;
	    params?: Record<string, string>;
	    data?: string;
	    error?: string;
	    statusCode?: number;

	    static createFrom(source: any = {}) {
	        return new APIResponse_string_(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.params = source["params"];
	        this.data = source["data"];
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }
	}

}

export namespace vault {

	export class SetupRequest {
	    masterPassword: string;

	    static createFrom(source: any = {}) {
	        return new SetupRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.masterPassword = source["masterPassword"];
	    }
	}
	export class UnlockRequest {
	    masterPassword: string;

	    static createFrom(source: any = {}) {
	        return new UnlockRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.masterPassword = source["masterPassword"];
	    }
	}
	export class VaultState {
	    isSetup: boolean;
	    isUnlocked: boolean;
	    autoLockSeconds: number;

	    static createFrom(source: any = {}) {
	        return new VaultState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isSetup = source["isSetup"];
	        this.isUnlocked = source["isUnlocked"];
	        this.autoLockSeconds = source["autoLockSeconds"];
	    }
	}

}

export namespace workspaces {

	export class CreateWorkspaceRequest {
	    projectId: string;
	    name: string;
	    color?: string;

	    static createFrom(source: any = {}) {
	        return new CreateWorkspaceRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	}
	export class UpdateWorkspaceRequest {
	    id: string;
	    name: string;
	    color?: string;

	    static createFrom(source: any = {}) {
	        return new UpdateWorkspaceRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	}
	export class WorkspaceDTO {
	    id: string;
	    projectId: string;
	    name: string;
	    color: string;
	    sortOrder: number;
	    createdAt: number;
	    updatedAt: number;
	    revision: string;

	    static createFrom(source: any = {}) {
	        return new WorkspaceDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.revision = source["revision"];
	    }
	}

}

