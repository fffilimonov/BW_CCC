package main

import (
    "fmt"
    "gopkg.in/gcfg.v1"
    "os"
    "strings"
    "time"
)

type ConfigT struct {
    Main struct {
        User string
        Password string
        Host string
        HTTPHost string
        HTTPPort string
        OCIPPort string
        Expires string
        Wraptime time.Duration
        TargetID []string
        Name []string
        Event []string
        CCID string
        CCEvent string
    }
}

type ConfigTlocal struct {
    Main struct {
        Owner string
    }
}

func ReadConfig(Configfile string) ConfigT {
    var Config ConfigT
    err := gcfg.ReadFileInto(&Config, Configfile)
    if err != nil {
        LogErr(err,"Config file is missing:", Configfile)
        os.Exit (1)
    }
    return Config
}

func ReadConfiglocal(Configfile string) ConfigTlocal {
    var Config ConfigTlocal
    err := gcfg.ReadFileInto(&Config, Configfile)
    if err != nil {
        LogErr(err,"Config file is missing:", Configfile)
        os.Exit (1)
    }
    return Config
}

func ConcatStr(sep string, args ... string) string {
    return strings.Join(args, sep)
}

func LogErr (err error,args ... string) {
    fmt.Fprint(os.Stderr,time.Now(),args,err,"\n")
}

func LogOut (log string) {
    fmt.Fprint(os.Stdout,log,"\n\n")
}

func Log2Out (args ... string) {
    fmt.Fprint(os.Stdout,args,"\n\n")
}

type lCalls struct {
    Addr string
    Time time.Time
}

type point struct {
    x int
    y int
}

func main() {
    larg:=len(os.Args)
    if larg < 3 {
        LogErr(nil,"no args")
        os.Exit (1)
    }
    var globalconf string = os.Args[1]
    var localconf string = os.Args[2]
    guiMain(globalconf,localconf)
}
