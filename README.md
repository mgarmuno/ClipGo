# ClipGo

Simple clipboard manager written in Go. It's still a work in progress.

![alt text](https://github.com/mgarmuno/ClipGo/blob/main/clipGoPic.png?raw=true)

# Dependencies

The clipboard history is shown via dmenu https://tools.suckless.org/dmenu/.

You will need to have installed https://github.com/kfish/xsel. If you want to use another program to
manipulate the clipboard you will have to change the source.

It's intended to be used with some program that captures clipboard when it changes.

You can use https://github.com/cdown/clipnotify:

    while ./clipnotify;
    do
        SelectedText="$(xsel)"
        CopiedText="$(xsel -b)"
        if [[ $CopiedText == $SelectedText ]]; then
            ~/dev/clipGo/clipGo add
        fi
    done


# Usage

It accepts three inputs as first parameter:

    add    -> Adds what is in the clipboard at the moment of the execution to the history.
    show   -> Shows the history and writes the selected entry to the clipboard.
    delete -> Deletes the entry from the history.

It's recommended to call delete and show through shortcuts in the DE or WM you use and add with a 
bash script as mentioned before.
