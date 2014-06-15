#include "gui.h"

#import <Cocoa/Cocoa.h>
#import "window.h"
#import "windowDelegate.h"
#import "glView.h"
#import "textView.h"
#import "blurView.h"
#import "progressView.h"
#import "popupView.h"
#import "subtitleView.h"
#import "startupView.h"
#import "app.h"

void initialize() {
    if (NSApp)
        return;

	[NSApplication sharedApplication];

    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    Application *appDelegate = [[Application alloc] init];
    [NSApp setDelegate:appDelegate];

	NSLog(@"initialized");
	//create memu bar
	id menubar = [[NSMenu new] autorelease];
    id appMenuItem = [[NSMenuItem new] autorelease];
    [menubar addItem:appMenuItem];
    [NSApp setMainMenu:menubar];
    id appMenu = [[NSMenu new] autorelease];

    NSMenuItem *searchSubtitleMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Search Subtitle"
        action:@selector(searchSubtitleMenuItemClick:) keyEquivalent:@""] autorelease];
    [searchSubtitleMenuItem setTarget: appDelegate];
    [appMenu addItem:searchSubtitleMenuItem];

    NSMenuItem *openFileMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Open..."
        action:@selector(openFileMenuItemClick:) keyEquivalent:@"o"] autorelease];
    [openFileMenuItem setTarget: appDelegate];
    [appMenu addItem:openFileMenuItem];

    id appName = [[NSProcessInfo processInfo] processName];
    id quitTitle = [@"Quit " stringByAppendingString:appName];
    id quitMenuItem = [[[NSMenuItem alloc] initWithTitle:quitTitle
        action:@selector(terminate:) keyEquivalent:@"q"] autorelease];
    [appMenu addItem:quitMenuItem];


    [appMenuItem setSubmenu:appMenu];
}
void hideMenuNSString(NSString* title) {
    NSLog(@"remove subtitle menu %@", title);

    NSMenu* menubar = [NSApp mainMenu];
    NSArray* menus = [menubar itemArray];
    for (NSMenuItem* menu in menus) {
        NSLog(@"compare %@ to %@", [menu title], title);
        if ([menu title] == title) {
            NSLog(@"remove menu item");
            [menubar removeItem:menu];
            break;
        }
    }
}
void hideSubtitleMenu() {
    hideMenuNSString(@"Subtitle");
}
void hideAudioMenu() {
    hideMenuNSString(@"Audio");
}
void initAudioMenu(void* wptr, char** names, int32_t* tags, int len, int selected) {
    hideAudioMenu();

    if (len > 0) {
        NSWindow* w = (NSWindow*)wptr;

        NSMenu *menubar = [NSApp mainMenu];
        NSMenuItem* audioMenuItem = [[NSMenuItem new] autorelease];
        [audioMenuItem setTitle:@"Audio"];
        [menubar addItem:audioMenuItem];
        NSMenu* audioMenu = [[NSMenu alloc] initWithTitle:@"Audio"];

        for (int i = 0; i < len; i++) {
            char* name = names[i];
            int tag = tags[i];
            NSMenuItem* item = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:name] 
                action:@selector(audioMenuItemClick:) keyEquivalent:@""];
            [item setTarget: w];
            [item setTag: tag];
            [audioMenu addItem:item];

            if (tag == selected) {
                [item setState: NSOnState];
            }
        }
        [audioMenuItem setSubmenu:audioMenu];
    }
}

void initSubtitleMenu(void* wptr, char** names, int32_t* tags, int len, int32_t selected1, int32_t selected2) {
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];

    hideSubtitleMenu();

    NSLog(@"len:%d", len);

    if (len > 0) {
        NSWindow* w = (NSWindow*)wptr;

        NSMenu* menubar = [NSApp mainMenu];
        NSMenuItem* subtitleMenuItem = [[NSMenuItem new] autorelease];
        [subtitleMenuItem setTitle:@"Subtitle"];
        [menubar addItem:subtitleMenuItem];
        NSMenu* subtitleMenu = [[NSMenu alloc] initWithTitle:@"Subtitle"];

        for (int i = 0; i < len; i++) {
            char* name = names[i];
            int tag = tags[i];
            NSMenuItem* item = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:name] 
                action:@selector(subtitleMenuItemClick:) keyEquivalent:@""];
            [item autorelease];
            [item setTarget: w];
            [item setTag: tag];
            [subtitleMenu addItem:item];

            if (tag == selected1) {
                [item setState: NSOnState];
            }

            if (tag == selected2) {
                [item setState: NSOnState];
            }
        }

        [subtitleMenuItem setSubmenu:subtitleMenu];
    }

    [pool drain];
}

void setWindowTitle(void* wptr, char* title) {
    Window* w = (Window*)wptr;
    [w setTitle:[NSString stringWithUTF8String:title]];
}

void setWindowSize(void* wptr, int width, int height) {
    Window* w = (Window*)wptr;

    NSRect frame = [w frame];
    frame.origin.y -= (height - frame.size.height)/2;
    frame.origin.x -= (width - frame.size.width)/2;
    frame.size = NSMakeSize(width, height);

    w->customAspectRatio = NSMakeSize(width, height);
    [w->glView setOriginalSize:NSMakeSize(width, height)];

    [w setFrame:frame display:YES animate:YES];
}

void* newWindow(char* title, int width, int height) {
	NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];

	initialize();

	Window* w = [[Window alloc] initWithTitle:[NSString stringWithUTF8String:title]
		width:width height:height];
	
    WindowDelegate* wd = (WindowDelegate*)[[WindowDelegate alloc] init];
	[w setDelegate:(id)wd];

	GLView* v = [[GLView alloc] initWithFrame2:NSMakeRect(0,0,width,height)];
    w->glView = v;
	// [w setContentView:v];
    // BlurView* topbv = [[BlurView alloc] initWithFrame:NSMakeRect(0, height-30,width,30)];
    // w->titlebarView = topbv;

    NSView* rv = [[w contentView] superview];

    v->frameView = rv;
    // [topbv setAutoresizingMask:NSViewWidthSizable];

    // TitlebarView* tbarv = [[TitlebarView alloc] initWithFrame:NSMakeRect(0, 0, width, 30)];
    // [topbv addSubview:tbarv];
    // [tbarv setAutoresizingMask:NSViewWidthSizable];



    // for (NSView* subv in rv.subviews) {
    //     if (subv != [w contentView]) {
    //         [subv removeFromSuperview];
    //         [topbv addSubview:subv];
    //     }
    // }
    // NSView *v0 = [rv.subviews objectAtIndex:0];
    // NSView *v1 = [rv.subviews objectAtIndex:1];
    // NSView *v2 = [rv.subviews objectAtIndex:2];
    // NSView *v3 = [rv.subviews objectAtIndex:3];

    // if (v0 != [w contentView]) {
    //     [v0 removeFromSuperview];
    //     [topbv addSubview:v0];
    // }
    // if (v1 != [w contentView]) {
    //     [v1 removeFromSuperview];
    //     [topbv addSubview:v1];
    // }
    // if (v2 != [w contentView]) {
    //     [v2 removeFromSuperview];
    //     [topbv addSubview:v2];
    // }
    // if (v3 != [w contentView]) {
    //     [v3 removeFromSuperview];
    //     [topbv addSubview:v3];
    // }

    // [tbarv setTitle:[NSString stringWithUTF8String:title]];

    NSView* roundView = [[NSView alloc] initWithFrame:NSMakeRect(0,0,width,height)];
    roundView.wantsLayer = YES;
    roundView.layer.masksToBounds = YES;
    roundView.layer.cornerRadius = 4.1;
    [roundView addSubview:v];

    [rv addSubview:roundView positioned:NSWindowBelow relativeTo:nil];

    // [rv addSubview:topbv];


    [roundView setFrame:NSMakeRect(0, 0, width, height)];
    [roundView setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    [v setFrame:NSMakeRect(0,0,width,height)];
    [v setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];


    [rv setWantsLayer:YES];
    rv.layer.cornerRadius=4.1;
    rv.layer.masksToBounds=YES;



    TextView* tv = [[TextView alloc] initWithFrame:NSMakeRect(0, 30, width, 0)];
    [v addSubview:tv];
    [tv setAutoresizingMask:NSViewWidthSizable];
    [v setTextView:tv];

    TextView* tv2 = [[TextView alloc] initWithFrame:NSMakeRect(0, 30, width, 0)];
    [v addSubview:tv2];
    [tv2 setAutoresizingMask:NSViewWidthSizable];
    [v setTextView2:tv2];

    BlurView* bv = [[BlurView alloc] initWithFrame:NSMakeRect(0,0,width,30)];
    [v addSubview:bv];
    [bv setAutoresizingMask:NSViewWidthSizable];

    ProgressView* pv = [[ProgressView alloc] initWithFrame:[bv frame]];
    [bv addSubview:pv];
    [pv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];

    [v setProgressView:pv];

    [w makeFirstResponder:v];
    v->win = w;

    // BlurView* bvPopup = [[BlurView alloc] initWithFrame:NSMakeRect(200,40,400,500)];
    // [v addSubview:bvPopup];
    // [bvPopup setAutoresizingMask:NSViewWidthSizable];

    // PopupView* ppv = [[PopupView alloc] initWithFrame:NSMakeRect(0,0,400,500)];
    // [bvPopup addSubview:ppv];
    // [ppv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];


    StartupView* sv = [[StartupView alloc] initWithFrame:[v frame]];
    [v addSubview:sv];
    [sv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];

    [v setStartupView:sv];
    // [sv setNeedsDisplay:NO];


    NSTimer *renderTimer = [NSTimer timerWithTimeInterval:1.0/100.0 
                            target:w
                          selector:@selector(timerTick:)
                          userInfo:nil
                           repeats:YES];

    [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                forMode:NSDefaultRunLoopMode];
    [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                forMode:NSEventTrackingRunLoopMode]; //Ensure timer fires during resize

	[pool drain];

	return w;
}

void showWindow(void* ptr) {
	[NSApp activateIgnoringOtherApps:YES];

	Window* w = (Window*)ptr;
	[w makeKeyAndOrderFront:nil];
}
void makeWindowCurrentContext(void*ptr) {
    Window* w = (Window*)ptr;
    [w makeCurrentContext];
}
void pollEvents() {
    NSAutoreleasePool* pool = [[NSAutoreleasePool alloc] init];
    [NSApp finishLaunching];

    while(YES) {
	    [pool drain];
		pool = [[NSAutoreleasePool alloc] init];

	    NSEvent* event = [NSApp nextEventMatchingMask:NSAnyEventMask
	                                        untilDate:[NSDate distantFuture]
	                                        inMode:NSDefaultRunLoopMode
	                                        dequeue:YES];
	    [NSApp sendEvent:event];
	}
	// [NSApp activateIgnoringOtherApps:YES];
    //[NSApp run];

    [pool drain];
}
void refreshWindowContent(void*wptr) {
	Window* w = (Window*)wptr;
	[w setContentViewNeedsDisplay:YES];
}

int getWindowWidth(void* ptr) {
    Window* w = (Window*)ptr;
    return (int)([w->glView frame].size.width);
}
int getWindowHeight(void* ptr) {
    Window* w = (Window*)ptr;
    return (int)([w->glView frame].size.height);
}
void showWindowProgress(void* ptr, char* left, char* right, double percent) {
    Window* w = (Window*)ptr;
    [w->glView showProgress:left right:right percent:percent];
}
void showWindowBufferInfo(void* ptr, char* speed, double percent) {
    Window* w = (Window*)ptr;
    [w->glView showBufferInfo:speed bufferPercent:percent];
}
void* showText(void* ptr, SubItem* items, int length, int position, double x, double y) {
    Window* w = (Window*)ptr;
    return [w->glView showText:items length:length position:position x:x y:y];
}
void hideText(void* ptrWin, void* ptrText) {
    Window* w = (Window*)ptrWin;
    [w->glView hideText:ptrText];
}
void windowHideStartupView(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView hideStartupView];
}
void windowShowStartupView(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView showStartupView];
}
void windowToggleFullScreen(void* ptr) {
    Window* w = (Window*)ptr;
    [w toggleFullScreen:nil];
}

void hideCursor(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView hideCursor];
    [w->glView hideProgress];
}

void showCursor(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView showCursor];
    [w->glView showProgress];
}

void *newDialog(char* title, int width, int height) {
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];

    NSPanel *dialog = [[NSPanel alloc] initWithContentRect:NSMakeRect(200.0, 200.0, 300, 200)
        styleMask:NSHUDWindowMask | NSClosableWindowMask | NSTitledWindowMask | NSUtilityWindowMask | NSResizableWindowMask
          backing:NSBackingStoreBuffered
            defer:YES];

    [dialog makeKeyAndOrderFront:nil];

    [dialog setTitle:[NSString stringWithUTF8String:title]];
    
    SubtitleView* sv = [[SubtitleView alloc] initWithFrame:NSMakeRect(0,0,dialog.frame.size.width,dialog.frame.size.height)];
    [dialog setContentView:sv];
    [sv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];


    [pool drain];

    return dialog;
}

CSize getScreenSize() {
    NSSize sz = [[NSScreen mainScreen] frame].size;
    CSize csz;
    csz.width = (int)sz.width;
    csz.height = (int)sz.height;
    return csz;
}