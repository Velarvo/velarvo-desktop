#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

static float gTLSize       = 12.0;
static float gTLTopPad     = 7.0;
static float gTLLeftPad    = 8.0;
static float gTLSpacing    = 20.0;

static void ApplyTrafficLightLayout(NSWindow *window) {
    if (!window) return;

    NSButton *close = [window standardWindowButton:NSWindowCloseButton];
    NSButton *mini  = [window standardWindowButton:NSWindowMiniaturizeButton];
    NSButton *zoom  = [window standardWindowButton:NSWindowZoomButton];

    if (!close || !mini || !zoom) return;

    NSView *container = close.superview;
    CGFloat containerHeight = container.frame.size.height;

    CGFloat y = containerHeight - gTLTopPad - gTLSize;

    [close setFrame:NSMakeRect(gTLLeftPad,                       y, gTLSize, gTLSize)];
    [mini  setFrame:NSMakeRect(gTLLeftPad + gTLSpacing,          y, gTLSize, gTLSize)];
    [zoom  setFrame:NSMakeRect(gTLLeftPad + gTLSpacing * 2.0,    y, gTLSize, gTLSize)];
}


@interface NSView (TrafficLightSwizzle)
- (void)swizzled_layout;
@end

@implementation NSView (TrafficLightSwizzle)
- (void)swizzled_layout {
    [self swizzled_layout]; 

    if ([NSStringFromClass([self class]) isEqualToString:@"NSTitlebarContainerView"]) {
        NSWindow *window = self.window;
        ApplyTrafficLightLayout(window);
    }
}
@end

static void SwizzleTitlebarLayout(void) {
    Class cls = NSClassFromString(@"NSTitlebarContainerView");
    if (!cls) return;

    SEL original = @selector(layout);
    SEL swizzled = @selector(swizzled_layout);

    Method origMethod    = class_getInstanceMethod(cls, original);
    Method swizzMethod   = class_getInstanceMethod([NSView class], swizzled);

    BOOL added = class_addMethod(cls,
                                 original,
                                 method_getImplementation(swizzMethod),
                                 method_getTypeEncoding(swizzMethod));
    if (added) {
        class_replaceMethod([NSView class],
                            swizzled,
                            method_getImplementation(origMethod),
                            method_getTypeEncoding(origMethod));
    } else {
        method_exchangeImplementations(origMethod, swizzMethod);
    }
}

void SetupTrafficLights(float size, float topPadding, float leftPadding, float spacing) {
    gTLSize     = size;
    gTLTopPad   = topPadding;
    gTLLeftPad  = leftPadding;
    gTLSpacing  = spacing;

    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        SwizzleTitlebarLayout();
    });

    dispatch_async(dispatch_get_main_queue(), ^{
        ApplyTrafficLightLayout([[NSApplication sharedApplication] mainWindow]);
    });
}