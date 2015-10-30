package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gdk"
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-gtk/gdkpixbuf"
    "github.com/fffilimonov/OCIP_go"
    "github.com/fffilimonov/XSI_go"
    "strings"
    "strconv"
    "time"
    "unsafe"
)

func guiMain (confglobal string,conflocal string) {
    var CallID string
    ch := make(chan string,100)
    Config := ReadConfig(confglobal)
    Configlocal := ReadConfiglocal(conflocal)
    owner:=Configlocal.Main.Owner

//prepare config for XSI
    var xsiConfig xsi.ConfigT
    xsiConfig.Main.User=Configlocal.Main.Owner
    xsiConfig.Main.Password=Configlocal.Main.Password
    xsiConfig.Main.Host=Config.Main.Host
    xsiConfig.Main.HTTPHost=Config.Main.HTTPHost
    xsiConfig.Main.HTTPPort=Config.Main.HTTPPort
    def := xsi.MakeDef(xsiConfig)

//start main client
    go clientMain(ch,Config)

//prepare config for OCI
    var ociConfig ocip.ConfigT
    ociConfig.Main.User=Configlocal.Main.Owner
    ociConfig.Main.Password=Configlocal.Main.Password
    ociConfig.Main.Host=Config.Main.Host
    ociConfig.Main.OCIPPort=Config.Main.OCIPPort
//set unavailable at start app
    ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Unavailable")
//prepare timer
    timer := time.NewTimer(time.Second)
    timer.Stop()

//init gthreads
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
    var p2r,p1r point
    var gdkwinr *gdk.Window
    p1r.x=-1
    p2r.y=-1
    var xr int = 0
    var yr int = 0
    var diffxr int = 0
    var diffyr int = 0
    pxr := &xr
    pyr := &yr

    resizearea := gtk.NewDrawingArea()
    resizearea.SetSizeRequest(10, 10)
    resizearea.Connect ("motion-notify-event", func(ctx *glib.CallbackContext) {
        if gdkwinr == nil {
            gdkwinr = resizearea.GetWindow()
        }
        argr := ctx.Args(0)
        mevr := *(**gdk.EventMotion)(unsafe.Pointer(&argr))
        var mtr gdk.ModifierType
        if mevr.IsHint != 0 {
            gdkwinr.GetPointer(&p2r.x, &p2r.y, &mtr)
        }
        if (gdk.EventMask(mtr)&gdk.BUTTON_PRESS_MASK) != 0 {
            if p1r.x!=-1 && p1r.y!=-1 {
                diffxr = p2r.x-p1r.x
                diffyr = p2r.y-p1r.y
                window.GetSize(pxr,pyr)
                window.Resize(xr+diffxr,yr+diffyr)
            }
        }
        p1r=p2r
    })
    resizearea.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.POINTER_MOTION_HINT_MASK | gdk.BUTTON_PRESS_MASK)) 

//menu
    menutable := gtk.NewTable(1, 8, true)
    menutable.Attach(movearea,0,6,0,1,gtk.FILL,gtk.FILL,0,0)
    menutable.Attach(btnhide,6,7,0,1,gtk.EXPAND,gtk.EXPAND,0,0)
    menutable.Attach(btnclose,7,8,0,1,gtk.EXPAND,gtk.EXPAND,0,0)

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
            tmptarget:=target
            tmp.Connect("clicked", func() {
                xsi.XSITransfer (xsiConfig,def,owner,CallID,tmptarget)
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

//calls
    table_cl := gtk.NewTable(2, 15, false)
    dlabel4 := make(map[uint]*gtk.Label)
    dlabel5 := make(map[uint]*gtk.Label)
    var i uint
    for i=0;i<15;i++{
        dlabel4[i] = gtk.NewLabel("")
        table_cl.Attach(dlabel4[i],0,1,i,i+1,gtk.FILL,gtk.FILL,1,1)
        dlabel5[i] = gtk.NewLabel("")
        table_cl.Attach(dlabel5[i],1,2,i,i+1,gtk.FILL,gtk.FILL,1,1)
    }

//tabs
    notebook := gtk.NewNotebook()
    notebook.AppendPage(table_ag, gtk.NewLabel("Агенты"))
    notebook.AppendPage(table_cl, gtk.NewLabel("Звонки"))

//add all to window
    vbox := gtk.NewVBox(false, 1)
    vbox.Add(menutable)
    vbox.Add(table)
    vbox.Add(notebook)
    vbox.Add(resizearea)

    swin.AddWithViewPort(vbox)
    window.Add(swin)
    window.ShowAll()

//main func for update
    go func() {
        for{
            select {
                case data := <-ch:
                    cinfo := strings.Split(strings.Trim(data, "\n"), ";")
//owner
                    if cinfo[0]==owner && cinfo[1]=="state" {
                        if cinfo[4] != "" {
                            CallID=cinfo[5]
                            gdk.ThreadsEnter()
                            owner2.SetLabel(strings.Trim(cinfo[4],"tel:"))
                            gdk.ThreadsLeave()
                        } else {
                            CallID=""
                            gdk.ThreadsEnter()
                            owner2.SetLabel("")
                            gdk.ThreadsLeave()
                        }
                        if cinfo[3] == "Available" {
                            gdk.ThreadsEnter()
                            owner3.SetFromPixbuf(pix["green"])
                            gdk.ThreadsLeave()
                        } else if cinfo[3] == "Wrap-Up" {
                            gdk.ThreadsEnter()
                            owner3.SetFromPixbuf(pix["yellow"])
                            gdk.ThreadsLeave()
                            timer.Reset(time.Second * Config.Main.Wraptime)
                        }else{
                            gdk.ThreadsEnter()
                            owner3.SetFromPixbuf(pix["grey"])
                            gdk.ThreadsLeave()
                        }
                    }
//CC q
                    if cinfo[0]==Config.Main.CCID && cinfo[1]=="state" {
                        if cinfo[6] != "" {
                            gdk.ThreadsEnter()
                            qlabel2.SetLabel(cinfo[6])
                            gdk.ThreadsLeave()
                        }
                    }
//CC calls
                    if cinfo[0]==Config.Main.CCID && cinfo[1]=="calls" {
                        if cinfo[3] != "" {
                            var i,j uint
                            j=2
                            for i=0;i<15;i++{
                                if cinfo[j] != "" {
                                    date,_:=strconv.Atoi(cinfo[j])
                                    date=date/1000
                                    j++
                                    Addr:=strings.Trim(cinfo[j],"tel:")
                                    j++
                                    Time:=time.Unix(int64(date),0)
                                    gdk.ThreadsEnter()
                                    tmp4:=dlabel4[i]
                                    tmp4.SetLabel(Time.Format(time.Stamp))
                                    tmp5:=dlabel5[i]
                                    tmp5.SetLabel(Addr)
                                    dlabel4[i]=tmp4
                                    dlabel5[i]=tmp5
                                    gdk.ThreadsLeave()
                                }
                            }
                        }
                    }
//Targets
                    if cinfo[0]!=owner && cinfo[0]!=Config.Main.CCID && cinfo[1]=="state" {
                        if cinfo[2]=="On-Hook" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel3[cinfo[0]]
                            tmp.SetFromPixbuf(pix["blank"])
                            dlabel3[cinfo[0]]=tmp
                            gdk.ThreadsLeave()
                        }
                        if cinfo[2]=="Off-Hook" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel3[cinfo[0]]
                            tmp.SetFromPixbuf(pix["call"])
                            dlabel3[cinfo[0]]=tmp
                            gdk.ThreadsLeave()
                        }
                        if cinfo[3] == "Available" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel2[cinfo[0]]
                            tmp.SetFromPixbuf(pix["green"])
                            dlabel2[cinfo[0]]=tmp
                            gdk.ThreadsLeave()
                        } else if cinfo[3] == "Wrap-Up" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel2[cinfo[0]]
                            tmp.SetFromPixbuf(pix["yellow"])
                            dlabel2[cinfo[0]]=tmp
                            gdk.ThreadsLeave()
                        }else{
                            gdk.ThreadsEnter()
                            tmp:=dlabel2[cinfo[0]]
                            tmp.SetFromPixbuf(pix["grey"])
                            dlabel2[cinfo[0]]=tmp
                            gdk.ThreadsLeave()
                        }
                    }
//timer for wrap-up
                case <-timer.C:
                    ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
            }
        }
    }()
    gtk.Main()
}
