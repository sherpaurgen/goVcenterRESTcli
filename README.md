# goVcenterRESTcli
simple go client for using vcenter rest api

First clone the repo and
cd /path/to/repo
go build

./vcenterapi -h
Usage of ./vcenterapi:
  -list
        Lists available virtual machines
  -start
        start vm700 vm701 #starts vms with name vm700 vm701
        
List vms & get the vmname eg. 
O/p order #vmid , vmName          ,PoweredState  , Memory(MB) , Num of cpu
./vcenterapi -list

vm-236,Normal_Windows_192.45.9.191,POWERED_OFF,mem:32768,cpu:16
vm-138,Normal_Windows_192.45.9.192,POWERED_OFF,POWERED_OFF,mem:32168,cpu:18
vm-182,Normal_Windows_192.45.9.193,POWERED_OFF,POWERED_ON,mem:4096,cpu:8



Example usage (when vms are already powered on)
main -start vm-5646 vm-69521
2022/03/01 00:19:00 [vm-5646 vm-69521]
2022/03/01 00:19:03 &{0x1120700 {0xc0000f09a0} 0x11d0e00} 400
2022/03/01 00:19:03 Problem starting vm-69521, already started state.
2022/03/01 00:19:03 &{0x1120700 {0xc0000f0b00} 0x11d0e00} 400
2022/03/01 00:19:03 Problem starting vm-5646, already started state.



