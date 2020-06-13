EncTool
=======

A tool for manipulating data in various encoding formats.


Usage
-----

```
Usage: enctool <command> [options]
Available commands:
    convert   : Convert between formats
    formats   : Print a list of supported formats
    print     : Print a document's structure.
    validate  : Validate a document.
```

Convert a [CTE](https://github.com/kstenerud/concise-encoding/blob/master/cte-condensed.md) document to [CBE](https://github.com/kstenerud/concise-encoding/blob/master/cbe-specification.md):

```
enctool convert -s=origdoc.cte -sf=cte -d=newdoc.cbe -df=cbe
```

Print a document's contents using 4 spaces indentation:

```
enctool print -f=document.cbe -fmt=cbe -i=4
```


License
-------

Copyright 2020 Karl Stenerud

Released under MIT license https://opensource.org/licenses/MIT

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
