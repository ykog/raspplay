/*
 * Raspberry piで遊ぶ
 */
package main

import (
  "fmt"
  "github.com/stianeikeland/go-rpio"
  "os"
  "time"
  "flag"
  linuxproc "github.com/c9s/goprocinfo/linux"
)

func getTicks() (total uint64,idle uint64,err error) {
  total = 0 
  idle = 0
  stat, err := linuxproc.ReadStat("/proc/stat")
  if err != nil {
    fmt.Println(err)
    return ;
  }
  s := stat.CPUStatAll 

  total = s.User + s.Nice + s.System + s.Idle + s.IOWait + s.IRQ + s.SoftIRQ + s.Steal + s.Guest + s.GuestNice ;
  idle = s.Idle ;
  return
}

func ledOnOff(pins []rpio.Pin,sw []bool) {
  for i:=0 ; i < len(pins) ; i++ {
    p := pins[i] ;
    if (sw[i]) {
      p.High()
    } else {
      p.Low()
    }
  }
}

func cpuStatMode() {

  pins := []rpio.Pin{ rpio.Pin(23), rpio.Pin(24), rpio.Pin(25) } ;
  sw := []bool{false,false,false}
  for i := 0 ; i < len(pins) ; i++ {
    pins[i].Output() ;
  }

  defer func() {
    for i := 0 ; i < len(pins) ; i++ {
      pins[i].Low() ;
    }
  }()


  // 最初のticksを取り出す
  lastTotal,lastIdle,err := getTicks()
  if err != nil {
    fmt.Println(err)
    os.Exit(1) 
  }

  // CPU使用率を求める繰返しに配流
  for true {
    time.Sleep(time.Second / 4)
    total,idle,err := getTicks()
    if err != nil {
      fmt.Println(err)
      os.Exit(1) 
    }

    totalTicks := float64(total - lastTotal)
    idleTicks := float64(idle - lastIdle)
    cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
    fmt.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)
    if (cpuUsage <= 15.0) {
      sw[0] = false
      sw[1] = false
      sw[2] = false
    } else if (cpuUsage <= 45.0) {
      sw[0] = true 
      sw[1] = false
      sw[2] = false
    } else if (cpuUsage <= 85.0) {
      sw[0] = true 
      sw[1] = true
      sw[2] = false
    } else {
      sw[0] = true 
      sw[1] = true
      sw[2] = true
    }
    ledOnOff(pins,sw)

    lastTotal = total ;
    lastIdle = idle ;
  }
}

func blinkMode(p int) {
  var pin = rpio.Pin(p)
  pin.Output()
  for x := 0 ; x < 20 ; x++ {
    pin.Toggle()
    time.Sleep(time.Second / 5)
  }
}

func main() {
  var p = flag.Int("p",0,"GPIO pin #")
  var c = flag.Bool("c",false,"CPU Stat mode") ;

  flag.Parse()

  if err := rpio.Open() ; err != nil {
    fmt.Println(err)
    os.Exit(1) 
  }
  defer rpio.Close()

  if (*c) {
    cpuStatMode() ;
    return ;
  } else if(*p > 0) {
    blinkMode(*p)
  }
}
