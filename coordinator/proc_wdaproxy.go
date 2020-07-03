package main

import (
  "fmt"
  log "github.com/sirupsen/logrus"
  "strconv"
  "strings"
  "time"
)

func restart_wdaproxy( devd *RunningDev ) {
    restart_proc_generic( devd, "wdaproxy" )
}
func wait_wdaup( devd *RunningDev ) {
    for {
        if devd.wda == true { break }
        time.Sleep( time.Second * 10 )
    }
}

func proc_wdaproxy( o ProcOptions, devEventCh chan<- DevEvent, temp bool ) {
    uuid := o.devd.uuid
    
    if temp {
      o.procName = "wdaproxytemp"
      o.noRestart = true
      o.noWait = false
    } else {
      o.procName = "wdaproxy"
      o.noWait = true
      o.noRestart = false
    }
    
    o.binary = "../wdaproxy" //o.config.BinPaths.WdaProxy
    o.startFields = log.Fields {
        "wdaPort": o.config.WDAProxyPort,
        "iosVersion": o.devd.iosVersion,
    }
    o.args = []string {
        "-p", strconv.Itoa(o.config.WDAProxyPort),
        "-q", strconv.Itoa(o.config.WDAProxyPort),
        "-d",
        "-W", ".",
        "-u", uuid,
        fmt.Sprintf("--iosversion=%s", o.devd.iosVersion),
    }
    o.startDir = o.config.WdaFolder
    
    o.stdoutHandler = func( line string, plog *log.Entry ) (bool) {
        if strings.Contains( line, "TEST EXECUTE FAILED" ) {
            plog.WithFields( log.Fields{
                "type": "wda_failed",
            } ).Error("WDA Failed")
            
            devEventCh <- DevEvent{
                action: 5,
                uuid: uuid,
            }
        }
        return true
    }
    
    devd := o.devd
    o.stderrHandler = func( line string, plog *log.Entry ) (bool) {
        if strings.Contains( line, "[WDA] successfully started" ) {
            /*plog.WithFields( log.Fields{
                "type": "wda_started",
            } ).Info("WDA Running")*/
            devd.lock.Lock()
            devd.wda = true
            devd.lock.Unlock()
            
            devEventCh <- DevEvent{
                action: 4,
                uuid: uuid,
            }
        }
        return true
    }
    o.onStop = func( devd *RunningDev ) {
        devd.lock.Lock()
        devd.wda = false
        devd.lock.Unlock()
    }
            
    proc_generic( o )
}