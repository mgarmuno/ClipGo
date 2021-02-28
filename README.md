# ClipGo

Simple clipboard manager written in Go.

# Dependencies

The clipboard history is shown via dmenu https://tools.suckless.org/dmenu/.

You will need to have installed https://github.com/kfish/xsel. If you want to use another program to
manipulate the clipboard you will have to change the source.

It's intended to be used with some program that captures clipboard when it changes.

You can use https://github.com/cdown/clipnotify:

    while clipnotify: do
        clipGo add
    done

# Usage

It accepts three inputs as first parameter:

    add    -> Adds what is in the clipboard at the moment of the execution to the history.
    show   -> Shows the history and writes the selected entry to the clipboard.
    delete -> Deletes the entry from the history.

It's recommended to call delete and show through shortcuts in the DE or WM you use and add with a 
shell script as mentioned before.