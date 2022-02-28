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
	"sync"
	"time"
	"flag"
)
type SessionData struct {
	VmwareApiSessionId string `json:"value"`
}
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
//waitgroup to keep track of the goroutines
var wg sync.WaitGroup

func powerOnVm(sessid string,vmname string,cli *http.Client,cred *Credential){
    // Endpoint https://{api_host}/api/vcenter/vm/{vm}/power?action=start
    hosturl:="https://"+ cred.Host + "/api/vcenter/vm/"+ vmname +"/power?action=start"
	req,err:=http.NewRequest("POST",hosturl,nil)
	req.Header.Add("vmware-api-session-id",sessid)
	resp,err := cli.Do(req)
	if err != nil {
		log.Fatal("Error %s", err)
	}
	defer resp.Body.Close()
	log.Print(resp.Body,resp.StatusCode)
	//informational messages reference
	//https://developer.vmware.com/apis/vsphere-automation/latest/vcenter/api/vcenter/vm/vm/poweractionstart/post/
	if resp.StatusCode == 204 {
		log.Print("Machine/s started successfully.")
	} else if resp.StatusCode == 400 {
		log.Printf("Problem starting %s, already in poweredOn state.",vmname)
	} else if resp.StatusCode == 404 {
		log.Printf("Problem starting %s, vm not found.",vmname)
	} else if resp.StatusCode == 500 {
		log.Printf("Problem starting %s, Virtualization Host error, please check logs.",vmname)
	} else if resp.StatusCode == 503 {
		log.Printf("Problem starting %s, com.vmware.vapi.std.errors.service_unavailable : if the system is unable to communicate with a service to complete the request.",vmname)
	}
	defer wg.Done()
}

func main(){

var err error
var loginurl string
var cred Credential
var hints string
hints="Create a user password file in your home dir eg. ~/.vmwarepass.json with contents similar to -> { \"host\":\"vm1.virtualdc.nu\",\"username\":\"john\",\"secret\":\"PqS4AqKjqkS#1\"}"
homedir,err := os.UserHomeDir()
homedir = homedir+"/.vmwarepass.json"
jsonFile,err:=os.Open(homedir)
	if err != nil{
		log.Print(hints)
		log.Fatal(err)
	}
	byteValue,_:=ioutil.ReadAll(jsonFile)
	defer jsonFile.Close()
	err = json.Unmarshal([]byte(byteValue),&cred) //parse username password json file into struct
	if err != nil {
		log.Print(hints)
		log.Fatal(err)
	}
	listvm:=flag.Bool("list",false,"Lists available virtual machines")
	startvm:=flag.Bool("start",false,"start vm700 vm701 #starts vms with name vm700 vm701")
	sessVal := &SessionData{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	sEnc := b64.StdEncoding.EncodeToString([]byte(cred.Username+":"+cred.Secret))
	loginurl = "https://"+cred.Host+"/rest/com/vmware/cis/session"
	flag.Parse()
	if *listvm {
		var allvmlist ColVmList
		//catch the sessionid and cliptr to reuse existing http client connection
		cliptr,sessVal := initializeConnection(loginurl,&cred,sEnc,sessVal)
		sessionid:=sessVal.VmwareApiSessionId
		allvmlist = getVmList(sessionid,cliptr,&cred)
		for _,val := range allvmlist.Value {
			//memory unit is in megabyte
			fmt.Printf("%s,%s,%s,mem:%s,cpu:%s\n",val.Vm,val.Name,val.Powerstat,strconv.Itoa(val.Mem),strconv.Itoa(val.Cpu))
		}
	}
	if *startvm {
		flag.Args()
		log.Print(flag.Args())
		if len(flag.Args())<1{
			log.Fatal("Please enter atleast one vm name")
		}
		cliptr,sessVal := initializeConnection(loginurl,&cred,sEnc,sessVal)
		sessionid:=sessVal.VmwareApiSessionId
		for v:=0;v<len(flag.Args());v++{
			wg.Add(1)
			go powerOnVm(sessionid,flag.Arg(v),cliptr,&cred)
		}
		wg.Wait()
	}

}

func initializeConnection(loginurl string,cred *Credential,sEnc string,sessVal *SessionData) (*http.Client,*SessionData) {
	cli:=http.Client{ Timeout: time.Second*10}
	req,err:=http.NewRequest("POST",loginurl,nil)
	req.SetBasicAuth(cred.Username, cred.Secret)
	req.Header.Add("Accept", `application/json`)
	req.Header.Add("Authorization","Authorization: Basic "+sEnc)
	resp,err := cli.Do(req)
	if err != nil {
		fmt.Printf("error %s", err)
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
	return &cli,sessVal
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


