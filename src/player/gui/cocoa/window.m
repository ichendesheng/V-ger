#import "window.h"

@implementation Window
- (id)initWithTitle:(NSString*)title width:(int)w height:(int)h  {
	unsigned int styleMask = NSTitledWindowMask | NSClosableWindowMask 
		| NSMiniaturizableWindowMask | NSResizableWindowMask;

	NSLog(@"%dx%d", w, h);

    self = [super initWithContentRect:NSMakeRect(0, 0, w, h)
    	styleMask:styleMask
    	backing:NSBackingStoreBuffered
      	defer:NO];

    [self setTitle:title];
    [self setContentAspectRatio:NSMakeSize(w, h)];
    [self setOpaque:YES];
    [self setHasShadow:YES];
    [self setContentMinSize:NSMakeSize(500, 500*h/w)];
    [self setAcceptsMouseMovedEvents:YES];
	[self setRestorable:NO];
    [self setCollectionBehavior:NSWindowCollectionBehaviorFullScreenPrimary];

    [self center];
    
    return self;
}

- (BOOL)canBecomeKeyWindow {
    return YES;
}
- (BOOL)isMovableByWindowBackground {
    return YES;
}
- (void)setContentViewNeedsDisplay:(BOOL)b {
	[[self contentView] setNeedsDisplay:b];
}
- (void)timerTick:(NSEvent *)event {
	onTimerTick((void*)self);
}
- (void)makeCurrentContext {
    [NSOpenGLContext clearCurrentContext];
    [[[self contentView] openGLContext] makeCurrentContext];
}
@end