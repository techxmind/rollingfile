# RollingFile [![Build Status](https://travis-ci.com/techxmind/rollingfile.svg?branch=main)](https://travis-ci.org/techxmind/rollingfile)
Implements io.WriteCloser, automatically rotate file with both file size and file lift time that you specified.

## USAGE
```
file, err := rollingfile.New(
    "current.data",     // active filename for writing
    MaxSize(1024*1000), // rotate when current file size reach 1024*1000 bytes
    MaxAge(3600),       // rotate when current file life time reach 3600 seconds
    Suffix("-old"),     // suffix used in rotated filename : current-20060102150405-old.data
)
```
