//go:build gui && darwin

package guiapp

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#include <stdlib.h>

void enableWebKitDevtools() {
    [[NSUserDefaults standardUserDefaults] setBool:YES forKey:@"WebKitDeveloperExtras"];
    [[NSUserDefaults standardUserDefaults] synchronize];
}

// findWKWebView recursively searches a view hierarchy for a WKWebView subclass.
static NSView* findWKWebView(NSView *view) {
    Class wkClass = NSClassFromString(@"WKWebView");
    if (wkClass && [view isKindOfClass:wkClass]) {
        return view;
    }
    for (NSView *sub in view.subviews) {
        NSView *found = findWKWebView(sub);
        if (found) return found;
    }
    return nil;
}

// openWebInspector enables developer extras and opens the WebKit inspector.
void openWebInspector() {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *win in [NSApp windows]) {
            NSView *wv = findWKWebView(win.contentView);
            if (!wv && win.contentViewController) {
                wv = findWKWebView(win.contentViewController.view);
            }
            if (!wv) continue;

            @try {
                id config = [wv valueForKey:@"configuration"];
                id prefs  = [config valueForKey:@"preferences"];
                [prefs setValue:@YES forKey:@"developerExtrasEnabled"];
            } @catch (NSException *e) {}

            SEL inspSel = NSSelectorFromString(@"_inspector");
            if ([wv respondsToSelector:inspSel]) {
                id inspector = [wv performSelector:inspSel];
                SEL showSel = NSSelectorFromString(@"show");
                if ([inspector respondsToSelector:showSel]) {
                    [inspector performSelector:showSel];
                }
            } else {
                SEL showSel = NSSelectorFromString(@"_showWebInspector");
                if ([wv respondsToSelector:showSel]) {
                    [wv performSelector:showSel];
                }
            }
            break;
        }
    });
}

void activateApp() {
    if ([NSThread isMainThread]) {
        [NSApp activateIgnoringOtherApps:YES];
    } else {
        dispatch_sync(dispatch_get_main_queue(), ^{
            [NSApp activateIgnoringOtherApps:YES];
        });
    }
}

// openDirectoryPanelSync opens an NSOpenPanel on the main thread and blocks
// until the user makes a selection. Returns a malloc'd C string (caller must free)
// or NULL if cancelled.
char* openDirectoryPanelSync(const char *title) {
    __block char *result = NULL;

    void (^showPanel)(void) = ^{
        [NSApp activateIgnoringOtherApps:YES];

        NSOpenPanel *panel = [NSOpenPanel openPanel];
        if (title && strlen(title) > 0) {
            panel.title = [NSString stringWithUTF8String:title];
        } else {
            panel.title = @"Select Folder";
        }
        panel.canChooseFiles = NO;
        panel.canChooseDirectories = YES;
        panel.allowsMultipleSelection = NO;
        panel.canCreateDirectories = YES;

        NSModalResponse response = [panel runModal];
        if (response == NSModalResponseOK && panel.URL != nil) {
            const char *path = [panel.URL.path UTF8String];
            result = strdup(path);
        }
    };

    if ([NSThread isMainThread]) {
        showPanel();
    } else {
        dispatch_sync(dispatch_get_main_queue(), showPanel);
    }

    return result;
}
*/
import "C"
import "unsafe"

func activateForDialog() {
	C.activateApp()
}

// EnableDevTools sets NSUserDefaults keys before WKWebView creation.
func EnableDevTools() {
	C.enableWebKitDevtools()
}

// PatchDevToolsAfterInit is unused but kept for interface compatibility.
func PatchDevToolsAfterInit() {}

// OpenWebInspector opens the WebKit inspector via private API.
func OpenWebInspector() {
	C.openWebInspector()
}

// selectDirectoryNative shows a native NSOpenPanel on the main thread.
func selectDirectoryNative(title string) string {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))

	cresult := C.openDirectoryPanelSync(ctitle)
	if cresult == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}
