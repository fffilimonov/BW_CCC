package main

import (
    "bufio"
    "fmt"
    "gopkg.in/gcfg.v1"
    "net"
    "os"
    "strings"
    "time"
)

type ConfigT struct {
    Main struct {
        Server string
        Port string
        Host string
        HTTPHost string
        HTTPPort string
        OCIPPort string
        Wraptime time.Duration
        TargetID []string
        Name []string
        CCID string
    }
}

type ConfigTlocal struct {
    Main struct {
        Owner string
        Password string
    }
}

func ConnectSrv (Config ConfigT) net.Conn {
    var dialer net.Dialer
    dialer.Timeout=time.Second
    chandesc, err := dialer.Dial("tcp", ConcatStr(":",Config.Main.Server,Config.Main.Port))
    if err != nil {
        LogErr(err,"serv dial")
        return nil
    } else {
        var send string = Config.Main.CCID
        for _,target := range Config.Main.TargetID {
            send=ConcatStr(" ",send,target)
        }
        fmt.Fprintf(chandesc,"%s",send)
        return chandesc
    }
}

func clientMain (ch chan string,Config ConfigT) {
    for {
        chandesc:=ConnectSrv(Config)
        if chandesc != nil {
            breader := bufio.NewReader(chandesc)
            for{
                str,err := breader.ReadString('\n')
                if err == nil {
                    ch<- string(str)
                } else {
                    chandesc.Close()
                    break
                }
            }
        }
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
