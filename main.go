package main

import (
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"flag"
)
type SessionData struct {
	VmwareApiSessionId string `json:"value"`
}
//{"memory_size_MiB":16384,"vm":"vm-10236","name":"Normal_Windows_ISO_2016_sab_20.15.19.100","power_state":"POWERED_OFF","cpu_count":8}
type Vm struct {
	Mem int `json:"memory_size_MiB"`
	Vm string `json:"vm"`
	Name string `json:"name"`
	Powerstat string `json:"power_state"`
	Cpu int `json:"cpu_count"`
}
type ColVmList struct {
	Value []Vm `json:"value"`
}
type Credential struct {
	Host string `json:"host"`
	Username string `json:"username"`
	Secret string `json:"secret"`
}

func getVmList(sessid string,cli *http.Client,cred *Credential) ColVmList {
	vms := ColVmList{}
	hosturl:="https://"+cred.Host+"/rest/vcenter/vm"
	req,err:=http.NewRequest("GET",hosturl,nil)
	req.Header.Add("vmware-api-session-id",sessid)
	resp,err := cli.Do(req)
	if err != nil {
		log.Fatal("Error %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(body),&vms)
	if err != nil {
		log.Fatal("Error %s", err)
	}
	return vms
}

func powerOnVm(sessid string,cli *http.Client,cred *Credential){
// Endpoint https://{api_host}/api/vcenter/vm/{vm}/power?action=start
	vmnameptr := flag.String("startvm", "", "specify vm name")
    hosturl:="https://"+cred.Host+"/api/vcenter/vm/"+ *vmnameptr +"/power?action=start"
	req,err:=http.NewRequest("POST",hosturl,nil)
	req.Header.Add("vmware-api-session-id",sessid)
	resp,err := cli.Do(req)
	if err != nil {
		log.Fatal("Error %s", err)
	}
	defer resp.Body.Close()
	log.Print(resp.Body,resp.StatusCode)
}



func main(){
var err error
var loginurl string
var cred Credential
homedir,err := os.UserHomeDir()
homedir = homedir+"/.vmwarepass.json"
jsonFile,err:=os.Open(homedir)
if err != nil{
	log.Print("Create a user password file in your home dir eg. ~/.vmwarepass.json with contents similar to -> { \"host\":\"vm1.virtualdc.nu\",\"username\":\"john\",\"secret\":\"PqS4AqKjqkS#1\"} ")
	log.Fatal(err)
}
byteValue,_:=ioutil.ReadAll(jsonFile)
defer jsonFile.Close()
err = json.Unmarshal([]byte(byteValue),&cred)
if err != nil{
	log.Print("Create a user password file in your home dir eg. ~/.vmwarepass.json with contents similar to -> { \"host\":\"vm1.virtualdc.nu\",\"username\":\"john\",\"secret\":\"PqS4AqKjqkS#1\"} ")
	log.Fatal(err)
	}

sessVal := &SessionData{}
http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
sEnc := b64.StdEncoding.EncodeToString([]byte(cred.Username+":"+cred.Secret))
loginurl = "https://"+cred.Host+"/rest/com/vmware/cis/session"
cli:=http.Client{ Timeout: time.Second*10}

req,err:=http.NewRequest("POST",loginurl,nil)
req.SetBasicAuth(cred.Username, cred.Secret)
req.Header.Add("Accept", `application/json`)
req.Header.Add("Authorization","Authorization: Basic "+sEnc)
resp,err := cli.Do(req)
if err != nil {
		fmt.Printf("error %s", err)
		return
	}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
	log.Fatal(err)
}
err = json.Unmarshal([]byte(string(body)),&sessVal)
if  err != nil{
	log.Fatal(err)
}
var allvmlist ColVmList
allvmlist = getVmList(sessVal.VmwareApiSessionId,&cli,&cred)
for _,val := range allvmlist.Value {
	fmt.Printf("%s,%s,%s,mem:%s,cpu:%s\n",val.Vm,val.Name,val.Powerstat,strconv.Itoa(val.Cpu),strconv.Itoa(val.Mem))
}

}


