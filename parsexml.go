package main

import (
    "encoding/xml"
)

type AcdST struct {
    Acd string `xml:"command>agentACDState"`
}

type QST struct {
    Qst string `xml:"command>numberOfCallsQueuedNow"`
}

type CallInfo struct {
    Target string `xml:"targetId"`
    Pers string `xml:"eventData>call>personality"`
    State string `xml:"eventData>call>state"`
    Hook string `xml:"eventData>hookStatus"`
    Addr string `xml:"eventData>call>remoteParty>address"`
    Aaddr string `xml:"eventData>queueEntry>remoteParty>address"`
    Atime string `xml:"eventData>queueEntry>removeTime"`
    Acount int `xml:"eventData>position"`
    CCstatus string `xml:"eventData>stateInfo>state"`
    CCstatuschanged string `xml:"eventData>agentStateInfo>state"`
    Etype string
    CallID string `xml:"eventData>call>callId"`
}

type Eventtype struct {
    Edata edata `xml:"eventData"`
}

type edata struct {
    Etype string `xml:"type,attr"`
}

func ParseData (data []byte) CallInfo {
    var callinfo CallInfo
    xml.Unmarshal(data, &callinfo)
    return callinfo
}

func ParseEdata (data []byte) Eventtype {
    var edata Eventtype
    xml.Unmarshal(data, &edata)
    return edata
}

func GetAcd (data []byte) string {
    var acd AcdST
    xml.Unmarshal(data, &acd)
    return acd.Acd
}

func GetQst (data []byte) string {
    var qst QST
    xml.Unmarshal(data, &qst)
    return qst.Qst
}
