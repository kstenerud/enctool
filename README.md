EncTool
=======

A tool for manipulating data in various encoding formats.



Prerelease
----------

Please note that this is still in pre-release, and although the most common functionality is in place, some less commonly used features are not yet implemented, and some edge cases may not be handled properly yet. Please be patient, and please file bug reports!

#### Prerelease Document Version

During the pre-release phase of the CE specification, all Concise Encoding documents will have a version of `0`. The tool will also accept version 1, but will always emit version 0.

#### How to report bugs

Quite often, the problem can be solved simply by upgrading to the latest Concise Encoding library like so:

* Go to the CE reference library repo's [commit list](https://github.com/kstenerud/go-concise-encoding/commits/master) and choose the latest commit (you'll see the short commit ID on the right-hand side).
* Go into the top level enctool dir in a shell.
* Use `go get` to fetch the library into enctool. For example `go get github.com/kstenerud/go-concise-encoding@884af95`.
* Tidy up the modules: `go mod tidy`
* Rebuild: `go build`

If updating the library doesn't fix it, file a bug report!

If updating the library does fix it, file a bug report telling me to update the module!



Building
--------

Type `go build` inside the top-level directory. It will create an executable file `enctool`.



Usage
-----

```
Usage: enctool <command> [options]
Available commands:
    convert   : Convert between formats
    formats   : Print a list of supported formats
    print     : Print a document's structure.
    validate  : Validate a document.

Use -h after a command to get a list of options.
```

Convert a [CTE](https://github.com/kstenerud/concise-encoding/blob/master/cte-specification.md) document to [CBE](https://github.com/kstenerud/concise-encoding/blob/master/cbe-specification.md):

```
enctool convert -s=origdoc.cte -sf=cte -d=newdoc.cbe -df=cbe
```

Print a document's contents using 4 spaces indentation:

```
enctool print -f=document.cbe -fmt=cbe -i=4
```

Enctool's convert mode also supports interpreting input as text-encoded byte values, using `-x` for hex encoded, or the more general `-t` which supports decimal format and `0xff` style hex. For example:

```
$ echo "81 00 ff" | ./enctool convert -x -sf cbe -df cte
c0
-1
$ echo "0x81, 0x00, 0xff" | ./enctool convert -t -sf cbe -df cte
c0
-1
$ echo "129, 0, 255" | ./enctool convert -t -sf cbe -df cte
c0
-1
$ echo "129 0 0xff" | ./enctool convert -t -sf cbe -df cte
c0
-1
```


License
-------

Copyright 2020 Karl Stenerud

Released under MIT license https://opensource.org/licenses/MIT

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
