This uses the python package `tkinter` to display it's interface, which is itself built on Tcl/Tk which Apple doesn't really install right.

Assuming you're using homebrew, you're going to need the following to get this to run on a mac

    brew install tcl-tk python-tk

The `tcl-tk` may not be necessary, but come on, you probably need it. Also you probably also should be running python via homebrew anyway.

You almost certainly need to open a new shell to get proper environment stuff after running the installation command above.