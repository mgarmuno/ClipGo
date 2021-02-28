# ClipGo

Simple clipboard manager written in Go.

# Dependencies

You will need to have installed https://github.com/kfish/xsel.

It's intended to be used with some program that captures clipboard when it changes.

You can use https://github.com/cdown/clipnotify:

    while clipnotify: do
        clipGo add
    done
