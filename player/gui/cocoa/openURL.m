#import "openURL.h"

@interface OpenURL ()

@end

@implementation OpenURL

- (void)windowDidLoad {
    [super windowDidLoad];
    
    // Implement this method to handle any initialization after your window controller's window has been loaded from its nib file.
}

- (IBAction)openClick:(id)sender {
        [[self window] close];
        onOpenFile([_txtURL.stringValue UTF8String]);
}

- (IBAction)cancelClick:(id)sender {
        [[self window] close];
}

- (BOOL)validateUrl:(NSString *)candidate {
	NSString *urlRegEx = @"\\s*(http|https)://\\S+\\s*";
	NSPredicate *urlTest = [NSPredicate predicateWithFormat:@"SELF MATCHES %@", urlRegEx]; 
	return [urlTest evaluateWithObject:candidate];
}

-(void)controlTextDidChange:(NSNotification *)obj {
	[_btnOpen setEnabled:[self validateUrl:_txtURL.stringValue]];
}
@end
