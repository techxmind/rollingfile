# RollingFile [![Build Status](https://travis-ci.com/techxmind/rollingfile.svg?branch=main)](https://travis-ci.org/techxmind/rollingfile)
Implements io.WriteCloser, automatically rotate file with both file size and file lift time that you specified.

## USAGE
```
file, err := rollingfile.New(
    // active filename for writing
    "current.data",

    // rotate when current file size reach 1024*1000 bytes
    rollingfile.MaxSize(1024*1000),

    // rotate when current file life time reach 3600 seconds
    rollingfile.MaxAge(3600),

    // suffix used in rotated filename : current-20060102150405-old.data
    rollingfile.Suffix("-old"),
)
```
