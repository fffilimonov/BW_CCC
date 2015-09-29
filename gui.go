package main

import (
    "container/ring"
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gdk"
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-gtk/gdkpixbuf"
    "github.com/fffilimonov/OCIP_go"
    "github.com/fffilimonov/XSI_go"
    "net"
    "strings"
    "strconv"
    "time"
    "unsafe"
)

func clientMain (guich chan CallInfo,Config xsi.ConfigT,owner string,def xsi.DefHead,Event []string,ccid string,ccevent string) {
/* chans */
    ch := make(chan string,100)
    datach := make(chan string,100)
    cCh := make(chan net.Conn)
/* start sybscription and reading to chan */
    go xsi.XSIresubscribe(Config,cCh,owner,Event,ccid,ccevent)
    go xsi.XSIread(ch,cCh)
/* handle reading */
    go xsi.XSImain(Config,def,ch,datach)
    for {
        select {
            case data := <-datach:
                cinfo := ParseData([]byte(data))
                einfo:=ParseEdata([]byte(data))
                cinfo.Etype=einfo.Edata.Etype
                guich<-cinfo
            default:
                time.Sleep(time.Millisecond*10)
        }
    }
}

func guiMain (confglobal string,conflocal string) {
    var CallID string
    ch := make(chan CallInfo,100)
    Config := ReadConfig(confglobal)
    Configlocal := ReadConfiglocal(conflocal)
    owner:=Configlocal.Main.Owner
//prepare config for XSI
    var xsiConfig xsi.ConfigT
    xsiConfig.Main.User=Config.Main.User
    xsiConfig.Main.Password=Config.Main.Password
    xsiConfig.Main.Host=Config.Main.Host
    xsiConfig.Main.HTTPHost=Config.Main.HTTPHost
    xsiConfig.Main.HTTPPort=Config.Main.HTTPPort
    xsiConfig.Main.Expires=Config.Main.Expires
    def := xsi.MakeDef(xsiConfig)
    go clientMain(ch,xsiConfig,owner,def,Config.Main.Event,Config.Main.CCID,Config.Main.CCEvent)
    list := ring.New(15)
//prepare config for OCI
    var ociConfig ocip.ConfigT
    ociConfig.Main.User=Config.Main.User
    ociConfig.Main.Password=Config.Main.Password
    ociConfig.Main.Host=Config.Main.Host
    ociConfig.Main.OCIPPort=Config.Main.OCIPPort

    timer := time.NewTimer(time.Second)
    timer.Stop()

    glib.ThreadInit(nil)
    gdk.ThreadsInit()
    gdk.ThreadsEnter()
    gtk.Init(nil)
//names
    names := make(map[string]string)
    for iter,target := range Config.Main.TargetID {
        names[target]=Config.Main.Name[iter]
    }
//icons to pixbuf map
    pix := make(map[string]*gdkpixbuf.Pixbuf)
    im_call := gtk.NewImageFromFile("ico/Call-Ringing-48.ico")
    pix["call"]=im_call.GetPixbuf()
    im_blank := gtk.NewImageFromFile("ico/Empty-48.ico")
    pix["blank"]=im_blank.GetPixbuf()
    im_green := gtk.NewImageFromFile("ico/Green-ball-48.ico")
    pix["green"]=im_green.GetPixbuf()
    im_grey := gtk.NewImageFromFile("ico/Grey-ball-48.ico")
    pix["grey"]=im_grey.GetPixbuf()
    im_yellow := gtk.NewImageFromFile("ico/Yellow-ball-48.ico")
    pix["yellow"]=im_yellow.GetPixbuf()

    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.SetTitle("Call Center")
    window.SetIcon(pix["call"])
    window.SetPosition(gtk.WIN_POS_CENTER)
    window.SetSizeRequest(350, 500)
    window.SetDecorated(false)
    window.SetResizable(true)
    window.Connect("destroy", gtk.MainQuit)

    swin := gtk.NewScrolledWindow(nil, nil)
    swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
    swin.SetShadowType(gtk.SHADOW_IN)

//owner
    owner1 := gtk.NewLabel(names[owner])
    owner2 := gtk.NewLabel("")
    owner3 := gtk.NewImage()
//qstatus
    qlabel1:=gtk.NewLabel("В очереди:")
    qlabel2:=gtk.NewLabel("")
//buttons
    b_av := gtk.NewButtonWithLabel("Доступен")
    b_av.SetCanFocus(false)
    b_av.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
        })
    b_un := gtk.NewButtonWithLabel("Недоступен")
    b_un.SetCanFocus(false)
    b_un.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Unavailable")
        })
    b_wr := gtk.NewButtonWithLabel("Дообработка")
    b_wr.SetCanFocus(false)
    b_wr.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Wrap-Up")
        })
//main table
    table := gtk.NewTable(3, 3, false)
    table.Attach(owner1,0,1,0,1,gtk.FILL,gtk.FILL,1,1)
    table.Attach(owner3,1,2,0,1,gtk.FILL,gtk.FILL,1,1)
    table.Attach(owner2,2,3,0,1,gtk.FILL,gtk.FILL,1,1)
    table.Attach(b_av,0,1,1,2,gtk.FILL,gtk.FILL,1,1)
    table.Attach(b_un,1,2,1,2,gtk.FILL,gtk.FILL,1,1)
    table.Attach(b_wr,2,3,1,2,gtk.FILL,gtk.FILL,1,1)
    table.Attach(qlabel1,0,1,2,3,gtk.FILL,gtk.FILL,1,1)
    table.Attach(qlabel2,1,2,2,3,gtk.FILL,gtk.FILL,1,1)

//menu buttons
    btnclose := gtk.NewToolButtonFromStock(gtk.STOCK_STOP)
    btnclose.SetCanFocus(false)
    btnclose.OnClicked(gtk.MainQuit)

    btnhide := gtk.NewToolButtonFromStock(gtk.STOCK_REMOVE)
    btnhide.SetCanFocus(false)
    btnhide.OnClicked(window.Iconify)
//move window
    var p2,p1 point
    var gdkwin *gdk.Window
    p1.x=-1
    p2.y=-1
    var x int = 0
    var y int = 0
    var diffx int = 0
    var diffy int = 0
    px := &x
    py := &y
    movearea := gtk.NewDrawingArea()
    movearea.Connect ("motion-notify-event", func(ctx *glib.CallbackContext) {
        if gdkwin == nil {
            gdkwin = movearea.GetWindow()
        }
        arg := ctx.Args(0)
        mev := *(**gdk.EventMotion)(unsafe.Pointer(&arg))
        var mt gdk.ModifierType
        if mev.IsHint != 0 {
            gdkwin.GetPointer(&p2.x, &p2.y, &mt)
        }
        if (gdk.EventMask(mt)&gdk.BUTTON_PRESS_MASK) != 0 {
            if p1.x!=-1 && p1.y!=-1 {
                window.GetPosition(px,py)
                diffx = p2.x-p1.x
                diffy = p2.y-p1.y
                window.Move(x+diffx,y+diffy)
            }
            p1.x=p2.x-diffx
            p1.y=p2.y-diffy
        } else {
            p1.x=-1
            p2.y=-1
        }
    })
    movearea.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.POINTER_MOTION_HINT_MASK | gdk.BUTTON_PRESS_MASK))
//resize window
    statusbar := gtk.NewStatusbar()
//menu
    menutable := gtk.NewTable(1, 8, true)
    menutable.Attach(movearea,0,6,0,1,gtk.FILL,gtk.FILL,0,0)
    menutable.Attach(btnhide,6,7,0,1,gtk.EXPAND,gtk.EXPAND,0,0)
    menutable.Attach(btnclose,7,8,0,1,gtk.EXPAND,gtk.EXPAND,0,0)

    notebook := gtk.NewNotebook()
//agents
    dlabel1 := make(map[string]*gtk.Label)
    dlabel2 := make(map[string]*gtk.Image)
    dlabel3 := make(map[string]*gtk.Image)
    b_tr := make(map[string]*gtk.Button)

    var count uint = 0
    for _,target := range Config.Main.TargetID {
        if target != owner {
            count=count+1
            dlabel1[target] = gtk.NewLabel(names[target])
            dlabel2[target] = gtk.NewImage()
            dlabel3[target] = gtk.NewImage()
            tmp := gtk.NewButtonWithLabel("Перевод")
            tmp.SetCanFocus(false)
//dirty hack
            tmptarget:=target
            tmp.Connect("clicked", func() {
                xsi.XSITransfer (xsiConfig,def,owner,CallID,tmptarget)
                notebook.SetCurrentPage(0)
            })
            b_tr[target]=tmp
        }
    }

    table_ag := gtk.NewTable(4, count+1, false)
    var place uint = 0
    for _,target := range Config.Main.TargetID {
        if target != owner {
            place=place+1
            table_ag.Attach(dlabel1[target],0,1,place,place+1,gtk.FILL,gtk.FILL,1,1)
            table_ag.Attach(dlabel3[target],2,3,place,place+1,gtk.FILL,gtk.FILL,1,1)
            table_ag.Attach(dlabel2[target],1,2,place,place+1,gtk.FILL,gtk.FILL,1,1)
            table_ag.Attach(b_tr[target],3,4,place,place+1,gtk.FILL,gtk.FILL,1,1)
        }
    }

    table_cl := gtk.NewTable(2, 15, false)
    dlabel4 := make(map[uint]*gtk.Label)
    dlabel5 := make(map[uint]*gtk.Label)
    var i uint
    for i=0;i<uint(list.Len());i++{
        dlabel4[i] = gtk.NewLabel("")
        table_cl.Attach(dlabel4[i],0,1,i,i+1,gtk.FILL,gtk.FILL,1,1)
        dlabel5[i] = gtk.NewLabel("")
        table_cl.Attach(dlabel5[i],1,2,i,i+1,gtk.FILL,gtk.FILL,1,1)
    }

    notebook.AppendPage(table_cl, gtk.NewLabel("Звонки"))
    notebook.AppendPage(table_ag, gtk.NewLabel("Агенты"))

//refresh on switch
    notebook.Connect("switch-page", func(){
        if notebook.GetCurrentPage() == 0 {
            for _,target := range Config.Main.TargetID {
                if target != owner {
                    hook:=xsi.XSIGetHook(xsiConfig,def,target)
                    if hook == "Off-Hook" {
                        tmp:=dlabel3[target]
                        tmp.SetFromPixbuf(pix["call"])
                        dlabel3[target]=tmp
                    }else{
                        tmp:=dlabel3[target]
                        tmp.SetFromPixbuf(pix["blank"])
                        dlabel3[target]=tmp
                    }
                    acdstatus:=GetAcd([]byte(ocip.OCIPsend(ociConfig,"UserCallCenterGetRequest19",ConcatStr("","userId=",target))))
                    if acdstatus == "Available" {
                        tmp:=dlabel2[target]
                        tmp.SetFromPixbuf(pix["green"])
                        dlabel2[target]=tmp
                    }else if acdstatus == "Wrap-Up" {
                        tmp:=dlabel2[target]
                        tmp.SetFromPixbuf(pix["yellow"])
                        dlabel2[target]=tmp
                    }else {
                        tmp:=dlabel2[target]
                        tmp.SetFromPixbuf(pix["grey"])
                        dlabel2[target]=tmp
                    }
                }
            }
        }
    })

    vbox := gtk.NewVBox(false, 1)
    vbox.Add(menutable)
    vbox.Add(table)
    vbox.Add(notebook)
    vbox.Add(statusbar)

    swin.AddWithViewPort(vbox)
    window.Add(swin)
    window.ShowAll()
    var qcount int = 0
    go func() {
        for{
            select {
                case cinfo := <-ch:
                    if cinfo.Target==owner {
                        if cinfo.Pers == "Terminator" && cinfo.State == "Alerting" {
                            gdk.ThreadsEnter()
                            owner2.SetLabel(strings.Trim(cinfo.Addr,"tel:"))
                            CallID=cinfo.CallID
                            gdk.ThreadsLeave()
                        }
                        if cinfo.Pers == "Terminator" && cinfo.State == "Released" {
                            gdk.ThreadsEnter()
                            CallID=""
                            owner2.SetLabel("")
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatus != "" || cinfo.CCstatuschanged != "" {
                            var ccstatus string
                            if cinfo.CCstatus != "" {
                                ccstatus=cinfo.CCstatus
                            }
                            if cinfo.CCstatuschanged != "" {
                                ccstatus=cinfo.CCstatuschanged
                            }
                            gdk.ThreadsEnter()
                                if ccstatus == "Available" {
                                    owner3.SetFromPixbuf(pix["green"])
                                } else if ccstatus == "Wrap-Up" {
                                    owner3.SetFromPixbuf(pix["yellow"])
                                    timer.Reset(time.Second * Config.Main.Wraptime)
                                }else{
                                    owner3.SetFromPixbuf(pix["grey"])
                                }
                            gdk.ThreadsLeave()
                        }
                    }
                    if cinfo.Etype=="xsi:ACDCallAddedEvent" {
                        qcount=cinfo.Acount
                        gdk.ThreadsEnter()
                        qlabel2.SetLabel(strconv.Itoa(qcount))
                        gdk.ThreadsLeave()
                    }
                    if cinfo.Etype=="xsi:ACDCallOfferedToAgentEvent" {
                        if qcount > 0 {
                            qcount--
                            gdk.ThreadsEnter()
                            qlabel2.SetLabel(strconv.Itoa(qcount))
                            gdk.ThreadsLeave()
                        }
                    }
                    if cinfo.Etype=="xsi:ACDCallAbandonedEvent" {
                        if qcount > 0 {
                            qcount--
                        }
                        date,_:=strconv.Atoi(cinfo.Atime)
                        date=date/1000
                        var tmp lCalls
                        tmp.Addr=strings.Trim(cinfo.Aaddr,"tel:")
                        tmp.Time=time.Unix(int64(date),0)
                        list.Value = tmp
                        list = list.Next()
                        var i uint
                        for i=0;i<uint(list.Len());i++{
                            list = list.Prev()
                            if list.Value != nil {
                                tmp:=list.Value.(lCalls)
                                gdk.ThreadsEnter()
                                qlabel2.SetLabel(strconv.Itoa(qcount))
                                tmp4:=dlabel4[i]
                                tmp4.SetLabel(tmp.Time.Format(time.Stamp))
                                tmp5:=dlabel5[i]
                                tmp5.SetLabel(tmp.Addr)
                                dlabel4[i]=tmp4
                                dlabel5[i]=tmp5
                                gdk.ThreadsLeave()
                            }
                        }
                    }
                case <-timer.C:
                    ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
                default:
                    time.Sleep(time.Millisecond*10)
            }
        }
    }()
    gtk.Main()
}
