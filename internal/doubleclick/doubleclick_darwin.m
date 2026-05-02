#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import <objc/runtime.h>

@interface DoubleClickMessageHandler : NSObject <WKScriptMessageHandler>
@end

@implementation DoubleClickMessageHandler
- (void)userContentController:(WKUserContentController *)userContentController 
      didReceiveScriptMessage:(WKScriptMessage *)message {
    if ([message.name isEqualToString:@"doubleClickMaximize"]) {
        dispatch_async(dispatch_get_main_queue(), ^{
            NSWindow *window = [[NSApplication sharedApplication] mainWindow];
            if (!window) return;
            
            if ((window.styleMask & NSWindowStyleMaskFullScreen) == NSWindowStyleMaskFullScreen) {
                return;
            }
            
            [window zoom:nil];
        });
    }
}
@end

static DoubleClickMessageHandler *gHandler = nil;

void InjectDoubleClickMaximize(void) {
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        if (!gHandler) {
            gHandler = [[DoubleClickMessageHandler alloc] init];
        }
        
        Class webViewClass = NSClassFromString(@"WKWebView");
        if (!webViewClass) return;
        
        SEL originalSelector = @selector(initWithFrame:configuration:);
        Method originalMethod = class_getInstanceMethod(webViewClass, originalSelector);
        IMP originalIMP = method_getImplementation(originalMethod);
        
        IMP newIMP = imp_implementationWithBlock(^id(id self, CGRect frame, WKWebViewConfiguration *config) {
            if (config && config.userContentController && gHandler) {
                [config.userContentController addScriptMessageHandler:gHandler name:@"doubleClickMaximize"];
                
                NSString *script = @
                    "document.addEventListener('dblclick', function(e) {"
                    "    let el = e.target;"
                    "    while (el) {"
                    "        if (el.hasAttribute && el.hasAttribute('data-wails-drag')) {"
                    "            window.webkit.messageHandlers.doubleClickMaximize.postMessage('toggle');"
                    "            e.preventDefault();"
                    "            return;"
                    "        }"
                    "        const style = window.getComputedStyle(el);"
                    "        if (style.getPropertyValue('-webkit-app-region') === 'drag' ||"
                    "            style.getPropertyValue('--wails-draggable') === 'drag') {"
                    "            window.webkit.messageHandlers.doubleClickMaximize.postMessage('toggle');"
                    "            e.preventDefault();"
                    "            return;"
                    "        }"
                    "        el = el.parentElement;"
                    "    }"
                    "}, true);";
                
                WKUserScript *userScript = [[WKUserScript alloc]
                    initWithSource:script
                    injectionTime:WKUserScriptInjectionTimeAtDocumentEnd
                    forMainFrameOnly:YES];
                
                [config.userContentController addUserScript:userScript];
            }
            
            typedef id (*OriginalInitFunc)(id, SEL, CGRect, WKWebViewConfiguration *);
            return ((OriginalInitFunc)originalIMP)(self, originalSelector, frame, config);
        });
        
        method_setImplementation(originalMethod, newIMP);
    });
}
