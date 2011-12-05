symlinkit
========

`symlinkit` augments `ln`.

It traverses a directory-tree and does the following:

* if it encounters a directory, it creates the same directory at the target location
* if it encounters a file, it creates a symlink to that file at the target location

Application
-----------

I wrote this so I could maintain my own binaries easier.
I keep all the sources at `~/srcs`. Whenever I build something I set
the target (or `--prefix` when using autotools) to `~/builds/<appname>`.
Now I can use `symlinkit` to create a cumulative tree in `~/tree`. So when I want to uninstall
or upgrade a programm, I delete the build in `~/builds`, recompile and rerun `symlinkit`.
Very clean!
