## Album

Simple static server with optional basic authentication.

## Usage

### Install

    go get github.com/weaming/s
    
### Options

    Usage: s [options] ROOT
    The ROOT is the directory to be serve.

      -a	Whether need authorization.
      -l string
            Listen [host]:port, default bind to 0.0.0.0 (default ":8000")
      -n int
            The maximum number of files in each page. (default 20)
      -p string
            Basic authorization password (default "admin")
      -u string
            Basic authorization username (default "admin")
