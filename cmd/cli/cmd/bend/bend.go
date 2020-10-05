package bend

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
	rand "richardmace.co.uk/boxwallet/cmd/cli/cmd/bend/rand"
)

const (
	CAppName        string = "BoxWallet"
	CBWAppVersion   string = "0.33.0"
	CAppFilename    string = "boxwallet"
	CAppFilenameWin string = "boxwallet.exe"
	CAppLogfile     string = "boxwallet.log"

	CAppBinDir    string = "boxwallet"
	CAppBinDirWin string = "BoxWallet"

	// Divi-cli command constants
	cCommandGetBCInfo     string = "getblockchaininfo"
	cCommandGetWInfo      string = "getwalletinfo"
	cCommandMNSyncStatus1 string = "mnsync"
	cCommandMNSyncStatus2 string = "status"

	// Divi-cli wallet commands
	cCommandDisplayWalletAddress string = "getaddressesbyaccount" // ./divi-cli getaddressesbyaccount ""
	cCommandDumpHDInfo           string = "dumphdinfo"            // ./divi-cli dumphdinfo
	// CCommandEncryptWallet - Needed by dash command
	CCommandEncryptWallet  string = "encryptwallet"    // ./divi-cli encryptwallet “a_strong_password”
	cCommandRestoreWallet  string = "-hdseed="         // ./divid -debug-hdseed=the_seed -rescan (stop divid, rename wallet.dat, then run command)
	cCommandUnlockWallet   string = "walletpassphrase" // ./divi-cli walletpassphrase “password” 0
	cCommandUnlockWalletFS string = "walletpassphrase" // ./divi-cli walletpassphrase “password” 0 true
	cCommandLockWallet     string = "walletlock"       // ./divi-cli walletlock

	// Divid Responses
	cDiviDNotRunningError     string = "error: couldn't connect to server"
	cDiviDDIVIServerStarting  string = "DIVI server starting"
	cDividRespWalletEncrypted string = "wallet encrypted"

	cGoDiviExportPath         string = "export PATH=$PATH:"
	CUninstallConfirmationStr string = "Confirm"
	CSeedStoredSafelyStr      string = "Confirm"

	// CMinRequiredMemoryMB - Needed by install command
	CMinRequiredMemoryMB int = 920
	CMinRequiredSwapMB   int = 2048

	// Wallet Security Statuses - Should be types?
	CWalletStatusLocked      string = "locked"
	CWalletStatusUnlocked    string = "unlocked"
	CWalletStatusLockedAndSk string = "locked-anonymization"
	CWalletStatusUnEncrypted string = "unencrypted"

	cRPCUserStr     string = "rpcuser"
	cRPCPasswordStr string = "rpcpassword"

	cUtfTick     string = "\u2713"
	CUtfTickBold string = "\u2714"

	cCircProg1 string = "\u25F7"
	cCircProg2 string = "\u25F6"
	cCircProg3 string = "\u25F5"
	cCircProg4 string = "\u25F4"

	cUtfLock string = "\u1F512"

	cProg1 string = "|"
	cProg2 string = "/"
	cProg3 string = "-"
	cProg4 string = "\\"
	cProg5 string = "|"
	cProg6 string = "/"
	cProg7 string = "-"
	cProg8 string = "\\"
)

// APPType - either APPTCLI, APPTCLICompiled, APPTInstaller, APPTUpdater, APPTServer
type APPType int

const (
	// APPTCLI - e.g. boxdivi
	APPTCLI APPType = iota
	// APPTCLICompiled - e.g. cli
	APPTCLICompiled
	// APPTInstaller e.g. godivi-installer
	//APPTInstaller
	// APPTUpdater e.g. update-godivi
	APPTUpdater
	// APPTUpdaterCompiled e.g. updater
	APPTUpdaterCompiled
	// APPTServer e.g. boxdivis
	//APPTServer
	// APPTServerCompiled e.g. web
	//APPTServerCompiled
)

// ProjectType - To allow external to determine what kind of wallet we are working with
type ProjectType int

const (
	PTDivi ProjectType = iota
	PTPhore
	PTPIVX
	PTTrezarcoin
	PTFeathercoin
)

// OSType - either ostArm, ostLinux or ostWindows
type OSType int

const (
	// OSTArm - Arm
	OSTArm OSType = iota
	// OSTLinux - Linux
	OSTLinux
	// OSTWindows - Windows
	OSTWindows
)

// WEType = either wetUnencrypted, wetLocked, wetUnlocked, weUnlockedForStaking
type WEType int

const (
	WETUnencrypted WEType = iota
	WETLocked
	WETUnlocked
	WETUnlockedForStaking
	WETUnknown
)

type GenericRespStruct struct {
	Result string      `json:"result"`
	Error  interface{} `json:"error"`
	ID     string      `json:"id"`
}

type GetAddressesByAccountRespStruct struct {
	Result []string    `json:"result"`
	Error  interface{} `json:"error"`
	ID     string      `json:"id"`
}

type GetInfoRespStruct struct {
	Result struct {
		Version         string  `json:"version"`
		Protocolversion int     `json:"protocolversion"`
		Walletversion   int     `json:"walletversion"`
		Balance         float64 `json:"balance"`
		Blocks          int     `json:"blocks"`
		Timeoffset      int     `json:"timeoffset"`
		Connections     int     `json:"connections"`
		Proxy           string  `json:"proxy"`
		Difficulty      float64 `json:"difficulty"`
		Testnet         bool    `json:"testnet"`
		Moneysupply     float64 `json:"moneysupply"`
		Keypoololdest   int     `json:"keypoololdest"`
		Keypoolsize     int     `json:"keypoolsize"`
		UnlockedUntil   int     `json:"unlocked_until"`
		Paytxfee        float64 `json:"paytxfee"`
		Relayfee        float64 `json:"relayfee"`
		StakingStatus   string  `json:"staking status"`
		Errors          string  `json:"errors"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    string      `json:"id"`
}

type WalletInfoPhoreRespStruct struct {
	Result struct {
		Walletversion int     `json:"walletversion"`
		Balance       float64 `json:"balance"`
		Txcount       int     `json:"txcount"`
		Keypoololdest int     `json:"keypoololdest"`
		Keypoolsize   int     `json:"keypoolsize"`
		UnlockedUntil int     `json:"unlocked_until"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    string      `json:"id"`
}

type privateSeedStruct struct {
	Hdseed             string `json:"hdseed"`
	Mnemonic           string `json:"mnemonic"`
	Mnemonicpassphrase string `json:"mnemonicpassphrase"`
}

type listTransactions []struct {
	Account         string        `json:"account"`
	Address         string        `json:"address"`
	Category        string        `json:"category"`
	Amount          float64       `json:"amount"`
	Vout            int           `json:"vout"`
	Confirmations   int           `json:"confirmations"`
	Bcconfirmations int           `json:"bcconfirmations"`
	Blockhash       string        `json:"blockhash"`
	Blockindex      int           `json:"blockindex"`
	Blocktime       int           `json:"blocktime"`
	Txid            string        `json:"txid"`
	Walletconflicts []interface{} `json:"walletconflicts"`
	Time            int           `json:"time"`
	Timereceived    int           `json:"timereceived"`
}

type stakingStatusStruct struct {
	Validtime       bool `json:"validtime"`
	Haveconnections bool `json:"haveconnections"`
	Walletunlocked  bool `json:"walletunlocked"`
	Mintablecoins   bool `json:"mintablecoins"`
	Enoughcoins     bool `json:"enoughcoins"`
	Mnsync          bool `json:"mnsync"`
	StakingStatus   bool `json:"staking status"`
}

type usd2AUDRespStruct struct {
	Rates struct {
		AUD float64 `json:"AUD"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}

type usd2GBPRespStruct struct {
	Rates struct {
		GBP float64 `json:"GBP"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}

type walletResponseType int

const (
	wrType walletResponseType = iota
	wrtUnknown
	wrtAllOK
	wrtNotReady
	wrtStillLoading
)

type WalletFixType int

const (
	WFType WalletFixType = iota
	WFTReIndex
	WFTReSync
)

var gLastMNSyncStatus string = ""

// Ticker related variables
var gGetTickerInfoCount int
var gPricePerCoinAUD usd2AUDRespStruct
var gPricePerCoinGBP usd2GBPRespStruct
var gTicker DiviTickerStruct

func AddNodesDiviAlreadyExist() (bool, error) {
	chd, _ := GetCoinHomeFolder(APPTCLI)

	exists, err := StringExistsInFile("addnode=", chd+CDiviConfFile)
	if err != nil {
		return false, nil
	}
	if exists {
		return true, nil
	}
	return false, nil
}

func AddAddNodesIfRequired() error {
	doExist, err := AddNodesDiviAlreadyExist()
	if err != nil {
		return err
	}
	if !doExist {
		bwconf, err := GetConfigStruct("", false) //GetCLIConfStruct()
		if err != nil {
			return fmt.Errorf("unable to GetConfigStruct - %v", err)
		}
		switch bwconf.ProjectType {
		case PTDivi:
			chd, _ := GetCoinHomeFolder(APPTCLI)
			if err := os.MkdirAll(chd, os.ModePerm); err != nil {
				return fmt.Errorf("unable to make directory - %v", err)
			}
			addnodes, err := getAddNodes()
			if err != nil {
				return fmt.Errorf("unable to getAddNodes - %v", err)
			}

			sAddnodes := string(addnodes[:])
			if !strings.Contains(sAddnodes, "addnode") {
				return fmt.Errorf("unable to retrieve addnodes, please try again")
			}

			if err := WriteTextToFile(chd+CDiviConfFile, sAddnodes); err != nil {
				return fmt.Errorf("unable to write addnodes to file - %v", err)
			}

		case PTTrezarcoin:

		default:
			err = errors.New("unable to determine ProjectType")
		}
	}
	return nil
}

// ConvertBCVerification - Convert Blockchain verification progress
func ConvertBCVerification(verificationPG float64) string {
	var sProg string
	var fProg float64

	fProg = verificationPG * 100
	sProg = fmt.Sprintf("%.2f", fProg)

	return sProg
}

func findProcess(key string) (int, string, error) {
	pname := ""
	pid := 0
	err := errors.New("not found")
	ps, _ := ps.Processes()

	for i := range ps {
		if ps[i].Executable() == key {
			pid = ps[i].Pid()
			pname = ps[i].Executable()
			err = nil
			break
		}
	}
	return pid, pname, err
}

func getAddNodes() ([]byte, error) {
	addNodesClient := http.Client{
		Timeout: time.Second * 3, // Maximum of 3 secs
	}

	req, err := http.NewRequest(http.MethodGet, cDiviAddNodeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "boxwallet")

	res, getErr := addNodesClient.Do(req)
	if getErr != nil {
		return nil, err
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	return body, nil
}

// GetAppFileName - Returns the name of the app binary file e.g. boxwallet or boxwallet.exe
func GetAppFileName() (string, error) {
	switch runtime.GOOS {
	case "arm":
		return CAppFilename, nil
	case "linux":
		return CAppFilename, nil
	case "windows":
		return CAppFilenameWin, nil
	default:
		err := errors.New("unable to determine runtime.GOOS")
		return "", err
	}

	return "", nil
}

// // DoPrivKeyFile - Handles the private key
// func DoPrivKeyFile() error {
// 	dbf, err := GetAppsBinFolder()
// 	if err != nil {
// 		return fmt.Errorf("Unable to GetAppsBinFolder: %v", err)
// 	}

// 	// User specified that they wanted to save their private key in a file.
// 	s := getWalletSeedDisplayWarning() + `

// Storing your private key in a file is risky.

// Please confirm that you understand the risks: `
// 	yn := getYesNoResp(s)
// 	if yn == "y" {
// 		fmt.Println("\nRequesting private seed...")
// 		s, err := runDCCommand(dbf+cDiviCliFile, cCommandDumpHDinfo, "Waiting for wallet to respond. This could take several minutes...", 30)
// 		// cmd := exec.Command(dbf+cDiviCliFile, cCommandDumpHDinfo)
// 		// out, err := cmd.CombinedOutput()
// 		if err != nil {
// 			return fmt.Errorf("Unable to run command: %v - %v", dbf+cDiviCliFile+cCommandDumpHDinfo, err)
// 		}

// 		// Now store the info in file
// 		err = WriteTextToFile(dbf+cWalletSeedFileGoDivi, s)
// 		if err != nil {
// 			return fmt.Errorf("error writing to file %s", err)
// 		}
// 		fmt.Println("Now please store the privte seed file somewhere safe. The file has been saved to: " + dbf + cWalletSeedFileGoDivi)
// 	}
// 	return nil
// }

// func doWalletAdressDisplay() error {
// 	err := gwc.RunCoinDaemon(false)
// 	if err != nil {
// 		return fmt.Errorf("Unable to RunCoinDaemon: %v ", err)
// 	}

// 	dbf, err := gwc.GetAppsBinFolder()
// 	if err != nil {
// 		return fmt.Errorf("Unable to GetAppsBinFolder: %v", err)
// 	}
// 	// Display wallet public address

// 	fmt.Println("\nRequesting wallet address...")
// 	s, err := RunDCCommandWithValue(dbf+cDiviCliFile, cCommandDisplayWalletAddress, `""`, "Waiting for wallet to respond. This could take several minutes...", 30)
// 	if err != nil {
// 		return fmt.Errorf("Unable to run command: %v - %v", dbf+cDiviCliFile+cCommandDisplayWalletAddress, err)
// 	}

// 	fmt.Println("\nWallet address received...")
// 	fmt.Println("")
// 	println(s)

// 	return nil
// }

//func getBlockchainInfo() (blockChainInfo, error) {
//	bci := blockChainInfo{}
//	dbf, _ := gwc.GetAppsBinFolder(gwc.APPTServer)
//
//	cmdBCInfo := exec.Command(dbf+gwc.CDiviCliFile, cCommandGetBCInfo)
//	out, _ := cmdBCInfo.CombinedOutput()
//	err := json.Unmarshal([]byte(out), &bci)
//	if err != nil {
//		return bci, err
//	}
//	return bci, nil
//}

// GetCoinDaemonFilename - Return the coin daemon file name e.g. divid
func GetCoinDaemonFilename(at APPType) (string, error) {
	var pt ProjectType
	switch at {
	case APPTCLI:
		conf, err := GetConfigStruct("", false)
		if err != nil {
			return "", err
		}
		pt = conf.ProjectType
	default:
		err := errors.New("unable to determine AppType")
		return "", err
	}

	switch pt {
	case PTDivi:
		return CDiviDFile, nil
	case PTFeathercoin:
		return CFeathercoinDFile, nil
	case PTPhore:
		return CPhoreDFile, nil
	case PTPIVX:
		return CPIVXDFile, nil
	case PTTrezarcoin:
		return CTrezarcoinDFile, nil
	default:
		err := errors.New("unable to determine ProjectType")
		return "", err
	}

	return "", nil
}

// GetCoinName - Returns the name of the coin e.g. Divi
func GetCoinName(at APPType) (string, error) {
	var pt ProjectType
	switch at {
	case APPTCLI:
		conf, err := GetConfigStruct("", false)
		if err != nil {
			return "", err
		}
		pt = conf.ProjectType
	default:
		err := errors.New("unable to determine AppType")
		return "", err
	}

	switch pt {
	case PTDivi:
		return CCoinNameDivi, nil
	case PTFeathercoin:
		return CCoinNameFeathercoin, nil
	case PTPhore:
		return CCoinNamePhore, nil
	case PTPIVX:
		return CCoinNamePIVX, nil
	case PTTrezarcoin:
		return CCoinNameTrezarcoin, nil
	default:
		err := errors.New("unable to determine ProjectType")
		return "", err
	}

	return "", nil
}

// GetAppsBinFolder - Returns the directory of where the BoxWallet binary files are stored
func GetAppsBinFolderOld() (string, error) {
	var s string
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	hd := u.HomeDir
	if runtime.GOOS == "windows" {
		// add the "appdata\roaming" part.
		s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(CAppBinDirWin)
	} else {
		s = AddTrailingSlash(hd) + AddTrailingSlash(CAppBinDir)
	}
	return s, nil
}

// GetCoinHomeFolder - Returns the home folder for the coin e.g. .divi
func GetCoinHomeFolder(at APPType) (string, error) {
	var pt ProjectType
	switch at {
	case APPTCLI:
		conf, err := GetConfigStruct("", false)
		if err != nil {
			return "", err
		}
		pt = conf.ProjectType
	default:
		err := errors.New("unable to determine AppType")
		return "", err
	}

	var s string
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	hd := u.HomeDir
	if runtime.GOOS == "windows" {
		// add the "appdata\roaming" part.
		switch pt {
		case PTDivi:
			s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(CDiviHomeDirWin)
		case PTFeathercoin:
			s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(CFeathercoinHomeDirWin)
		case PTPhore:
			s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(CPhoreHomeDirWin)
		case PTPIVX:
			s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(cPIVXHomeDirWin)
		case PTTrezarcoin:
			s = AddTrailingSlash(hd) + "appdata\\roaming\\" + AddTrailingSlash(CTrezarcoinHomeDirWin)
		default:
			err = errors.New("unable to determine ProjectType")

		}
	} else {
		switch pt {
		case PTDivi:
			s = AddTrailingSlash(hd) + AddTrailingSlash(CDiviHomeDir)
		case PTFeathercoin:
			s = AddTrailingSlash(hd) + AddTrailingSlash(CFeathercoinHomeDir)
		case PTPhore:
			s = AddTrailingSlash(hd) + AddTrailingSlash(CPhoreHomeDir)
		case PTPIVX:
			s = AddTrailingSlash(hd) + AddTrailingSlash(cPIVXHomeDir)
		case PTTrezarcoin:
			s = AddTrailingSlash(hd) + AddTrailingSlash(CTrezarcoinHomeDir)
		default:
			err = errors.New("unable to determine ProjectType")

		}
	}
	return s, nil
}

//func getMNSyncStatus() (mnSyncStatus, error) {
//	// gdConfig, err := getConfStruct("./")
//	// if err != nil {
//	// 	log.Print(err)
//	// }
//
//	mnss := mnSyncStatus{}
//	dbf, _ := gwc.GetAppsBinFolder(gwc.APPTServer)
//
//	cmdMNSS := exec.Command(dbf+gwc.CDiviCliFile, cCommandMNSyncStatus1, cCommandMNSyncStatus2)
//	out, _ := cmdMNSS.CombinedOutput()
//	err := json.Unmarshal([]byte(out), &mnss)
//	if err != nil {
//		return mnss, err
//	}
//	return mnss, nil
//}

func getNextProgMNIndicator(LIndicator string) string {
	if LIndicator == cProg1 {
		gLastMNSyncStatus = cProg2
		return cProg2
	} else if LIndicator == cProg2 {
		gLastMNSyncStatus = cProg3
		return cProg3
	} else if LIndicator == cProg3 {
		gLastMNSyncStatus = cProg4
		return cProg4
	} else if LIndicator == cProg4 {
		gLastMNSyncStatus = cProg5
		return cProg5
	} else if LIndicator == cProg5 {
		gLastMNSyncStatus = cProg6
		return cProg6
	} else if LIndicator == cProg6 {
		gLastMNSyncStatus = cProg7
		return cProg7
	} else if LIndicator == cProg7 {
		gLastMNSyncStatus = cProg8
		return cProg8
	} else if LIndicator == cProg8 || LIndicator == "" {
		gLastMNSyncStatus = cProg1
		return cProg1
	} else {
		gLastMNSyncStatus = cProg1
		return cProg1
	}
}

// GetWalletAddress - Sends a "getaddressesbyaccount" to the daemon, and returns the result
func GetWalletAddress(cliConf *ConfStruct) (GetAddressesByAccountRespStruct, error) {
	var respStruct GetAddressesByAccountRespStruct

	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getaddressesbyaccount\",\"params\":[\"\"]}")
	req, err := http.NewRequest("POST", "http://"+cliConf.ServerIP+":"+cliConf.Port, body)
	if err != nil {
		return respStruct, err
	}
	req.SetBasicAuth(cliConf.RPCuser, cliConf.RPCpassword)
	req.Header.Set("Content-Type", "text/plain;")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return respStruct, err
	}
	defer resp.Body.Close()
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return respStruct, err
	}
	//s := string(bodyResp)
	//fmt.Println(s)
	err = json.Unmarshal(bodyResp, &respStruct)
	if err != nil {
		return respStruct, err
	}
	return respStruct, nil
}

// func getWalletAddress(attempts int) (string, error) {
// 	var err error
// 	var s string = "waiting for wallet."
// 	dbf, _ := gwc.GetAppsBinFolder(gwc.APPTServer)
// 	app := dbf + gwc.CDiviCliFile
// 	arg1 := cCommandDisplayWalletAddress
// 	arg2 := ""

// 	for i := 0; i < attempts; i++ {

// 		cmd := exec.Command(app, arg1, arg2)
// 		out, err := cmd.CombinedOutput()

// 		if err == nil {
// 			return string(out), err
// 		}

// 		fmt.Printf("\r"+s+" %d/"+strconv.Itoa(attempts), i+1)

// 		time.Sleep(3 * time.Second)
// 	}

// 	return "", err

// }

// func GetWalletInfo(dispProgress bool) (walletInfoStruct, walletResponseType, error) {
// 	wi := walletInfoStruct{}

// 	// Start the DiviD server if required...
// 	err := RunCoinDaemon(false)
// 	if err != nil {
// 		return wi, wrtUnknown, fmt.Errorf("Unable to RunDiviD: %v ", err)
// 	}

// 	dbf, err := gwc.GetAppsBinFolder(gwc.APPTServer)
// 	if err != nil {
// 		return wi, wrtUnknown, fmt.Errorf("Unable to GetAppsBinFolder: %v ", err)
// 	}

// 	for i := 0; i <= 4; i++ {
// 		cmd := exec.Command(dbf+gwc.CDiviCliFile, cCommandGetWInfo)
// 		var stdout bytes.Buffer
// 		cmd.Stdout = &stdout
// 		cmd.Run()
// 		if err != nil {
// 			return wi, wrtUnknown, err
// 		}

// 		outStr := string(stdout.Bytes())
// 		wr := getWalletResponse(outStr)

// 		// cmd := exec.Command(dbf+gwc.CDiviCliFile, cCommandGetWInfo)
// 		// out, err := cmd.CombinedOutput()
// 		// sOut := string(out)
// 		//wr := getWalletResponse(sOut)
// 		if wr == wrtAllOK {
// 			errUM := json.Unmarshal([]byte(outStr), &wi)
// 			if errUM == nil {
// 				return wi, wrtAllOK, err
// 			}
// 		}

// 		time.Sleep(1 * time.Second)
// 	}

// 	return wi, wrtUnknown, errors.New("Unable to retrieve wallet info")
// }

func GetPasswordToEncryptWallet() string {
	for i := 0; i <= 2; i++ {
		epw1 := ""
		prompt := &survey.Password{
			Message: "Please enter a password to encrypt your wallet",
		}
		survey.AskOne(prompt, &epw1)

		epw2 := ""
		prompt2 := &survey.Password{
			Message: "Now please re-enter your password",
		}
		survey.AskOne(prompt2, &epw2)
		if epw1 != epw2 {
			fmt.Print("\nThe passwords don't match, please try again...\n")
		} else {
			return epw1
		}
	}
	return ""
	//reader := bufio.NewReader(os.Stdin)
	//for i := 0; i <= 2; i++ {
	//	fmt.Print("\nPlease enter a password to encrypt your wallet: ")
	//	epw1, _ := reader.ReadString('\n')
	//	fmt.Print("\nNow please re-enter your password: ")
	//	epw2, _ := reader.ReadString('\n')
	//	if epw1 != epw2 {
	//		fmt.Print("\nThe passwords don't match, please try again...\n")
	//	} else {
	//		s := strings.ReplaceAll(epw1, "\n", "")
	//
	//		return s
	//	}
	//}
	//return ""
}

func GetWalletEncryptionPassword() string {
	pword := ""
	prompt := &survey.Password{
		Message: "Please enter your encrypted wallet password",
	}
	survey.AskOne(prompt, &pword)
	return pword
}

func GetWalletEncryptionResp() bool {
	ans := false
	prompt := &survey.Confirm{
		Message: `Your wallet is currently UNENCRYPTED!

It is *highly* recommended that you encrypt your wallet before proceeding any further.

Encrypt it now?:`,
	}
	survey.AskOne(prompt, &ans)
	return ans
}

func getWalletResponse(sOut string) walletResponseType {
	if sOut == "" {
		return wrtNotReady
	}

	if strings.Contains(sOut, "hdseed") {
		return wrtAllOK
	}

	if strings.Contains(sOut, "wallet") {
		return wrtAllOK
	}

	return wrtUnknown
}

// IsCoinDaemonRunning - Works out whether the coin Daemon is running e.g. divid
func IsCoinDaemonRunning() (bool, int, error) {
	var pid int
	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return false, pid, err
	}
	switch gwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			pid, _, err = findProcess(CDiviDFileWin)
		} else {
			pid, _, err = findProcess(CDiviDFile)
		}
	case PTFeathercoin:
		if runtime.GOOS == "windows" {
			pid, _, err = findProcess(CFeathercoinDFileWin)
		} else {
			pid, _, err = findProcess(CFeathercoinDFile)
		}
	case PTPhore:
		if runtime.GOOS == "windows" {
			pid, _, err = findProcess(CPhoreDFileWin)
		} else {
			pid, _, err = findProcess(CPhoreDFile)
		}
	case PTPIVX:
		if runtime.GOOS == "windows" {
			pid, _, err = findProcess(CPIVXDFileWin)
		} else {
			pid, _, err = findProcess(CPIVXDFile)
		}
	case PTTrezarcoin:
		if runtime.GOOS == "windows" {
			pid, _, err = findProcess(CTrezarcoinDFileWin)
		} else {
			pid, _, err = findProcess(CTrezarcoinDFile)
		}
	default:
		err = errors.New("unable to determine ProjectType")
	}

	if err == nil {
		return true, pid, nil //fmt.Printf ("Pid:%d, Pname:%s\n", pid, s)
	}
	return false, 0, err
}

// PopulateDaemonConfFile - Populates the divi.conf file
func PopulateDaemonConfFile() (rpcuser, rpcpassword string, err error) {

	bFileHasBeenBU := false
	bwconf, err := GetConfigStruct("", false)
	if err != nil {
		return "", "", fmt.Errorf("unable to GetConfigStruct - %v", err)
	}

	// Create the coins home folder if required...
	chd, _ := GetCoinHomeFolder(APPTCLI)
	if err := os.MkdirAll(chd, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("unable to make directory - %v", err)
	}

	// Create user and password variables
	var rpcu string
	var rpcpw string

	switch bwconf.ProjectType {
	case PTDivi:
		fmt.Println("Populating " + CDiviConfFile + " for initial setup...")

		// Add rpcuser info if required, or retrieve the existing one
		b, err := StringExistsInFile(cRPCUserStr+"=", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcu = "divirpc"
			if err := WriteTextToFile(chd+CDiviConfFile, cRPCUserStr+"="+rpcu); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcu, err = GetStringAfterStrFromFile(cRPCUserStr+"=", chd+CDiviConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add rpcpassword info if required, or retrieve the xisting one
		b, err = StringExistsInFile(cRPCPasswordStr+"=", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
				rpcpw = rand.String(20)
				if err := WriteTextToFile(chd+CDiviConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
					log.Fatal(err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, ""); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			rpcpw, err = GetStringAfterStrFromFile(cRPCPasswordStr+"=", chd+CDiviConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add daemon=1 info if required
		b, err = StringExistsInFile("daemon=1", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, "daemon=1"); err != nil {
					log.Fatal(err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, ""); err != nil {
					log.Fatal(err)
				}
			}
		}
		// Add server=1 info if required
		b, err = StringExistsInFile("server=1", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, "server=1"); err != nil {
					log.Fatal(err)
				}
			}
		}
		// Add rpcallowip= info if required
		b, err = StringExistsInFile("rpcallowip=", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, "rpcallowip=192.168.1.0/255.255.255.0"); err != nil {
					log.Fatal(err)
				}
			}
		}
		// Add rpcport= info if required
		b, err = StringExistsInFile("rpcport=", chd+CDiviConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CDiviConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
				if err := WriteTextToFile(chd+CDiviConfFile, "rpcport="+CDiviRPCPort); err != nil {
					log.Fatal(err)
				}
			}
		}
		return rpcu, rpcpw, nil
	case PTFeathercoin:
		fmt.Println("Populating " + CFeathercoinConfFile + " for initial setup...")

		// Add rpcuser info if required, or retrieve the existing one
		bNeedToWriteStr := true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile(cRPCUserStr+"=", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
				rpcu, err = GetStringAfterStrFromFile(cRPCUserStr+"=", chd+CFeathercoinConfFile)
				if err != nil {
					return "", "", fmt.Errorf("unable to search for text in file - %v", err)
				}
			}
		} else {
			if bNeedToWriteStr {
				rpcu = "feathercoinrpc"
				if err := WriteTextToFile(chd+CFeathercoinConfFile, cRPCUserStr+"="+rpcu); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Add rpcpassword info if required, or retrieve the existing one
		bNeedToWriteStr = true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile(cRPCPasswordStr+"=", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
				rpcpw, err = GetStringAfterStrFromFile(cRPCPasswordStr+"=", chd+CFeathercoinConfFile)
				if err != nil {
					return "", "", fmt.Errorf("unable to search for text in file - %v", err)
				}
			}
		} else {
			if bNeedToWriteStr {
				rpcpw = rand.String(20)
				if err := WriteTextToFile(chd+CFeathercoinConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
					log.Fatal(err)
				}
				if err := WriteTextToFile(chd+CFeathercoinConfFile, ""); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Add daemon=1 info if required
		bNeedToWriteStr = true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile("daemon=1", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
			}
		} else {
			if bNeedToWriteStr {
				if err := WriteTextToFile(chd+CFeathercoinConfFile, "daemon=1"); err != nil {
					log.Fatal(err)
				}
				if err := WriteTextToFile(chd+CFeathercoinConfFile, ""); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Add server=1 info if required
		bNeedToWriteStr = true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile("server=1", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
			}
		} else {
			if bNeedToWriteStr {
				if err := WriteTextToFile(chd+CFeathercoinConfFile, "server=1"); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Add rpcallowip= info if required
		bNeedToWriteStr = true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile("rpcallowip=", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
			}
		} else {
			if bNeedToWriteStr {
				if err := WriteTextToFile(chd+CFeathercoinConfFile, "rpcallowip=192.168.1.0/255.255.255.0"); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Add rpcport= info if required
		bNeedToWriteStr = true
		if FileExists(chd + CFeathercoinConfFile) {
			bStrFound, err := StringExistsInFile("rpcport=", chd+CFeathercoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
			if !bStrFound {
				// String not found
				if !bFileHasBeenBU {
					bFileHasBeenBU = true
					if err := BackupFile(chd, CFeathercoinConfFile, false); err != nil {
						return "", "", fmt.Errorf("unable to backup file - %v", err)
					}
				}
			} else {
				bNeedToWriteStr = false
			}
		} else {
			if bNeedToWriteStr {
				if err := WriteTextToFile(chd+CFeathercoinConfFile, "rpcport="+CFeathercoinRPCPort); err != nil {
					log.Fatal(err)
				}
			}
		}

		return rpcu, rpcpw, nil
	case PTPhore:
		fmt.Println("Populating " + CPhoreConfFile + " for initial setup...")

		// Add rpcuser info if required
		b, err := StringExistsInFile(cRPCUserStr+"=", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcu = "phorerpc"
			if err := WriteTextToFile(chd+CPhoreConfFile, cRPCUserStr+"="+rpcu); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcu, err = GetStringAfterStrFromFile(cRPCUserStr+"=", chd+CPhoreConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add rpcpassword info if required
		b, err = StringExistsInFile(cRPCPasswordStr+"=", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcpw = rand.String(20)
			if err := WriteTextToFile(chd+CPhoreConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, ""); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcpw, err = GetStringAfterStrFromFile(cRPCPasswordStr+"=", chd+CPhoreConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add daemon=1 info if required
		b, err = StringExistsInFile("daemon=1", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, "daemon=1"); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, ""); err != nil {
				log.Fatal(err)
			}
		}

		// Add server=1 info if required
		b, err = StringExistsInFile("server=1", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, "server=1"); err != nil {
				log.Fatal(err)
			}
		}
		// Add rpcallowip= info if required
		b, err = StringExistsInFile("rpcallowip=", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, "rpcallowip=192.168.1.0/255.255.255.0"); err != nil {
				log.Fatal(err)
			}
		}
		// Add rpcport= info if required
		b, err = StringExistsInFile("rpcport=", chd+CPhoreConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPhoreConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPhoreConfFile, "rpcport=11772"); err != nil {
				log.Fatal(err)
			}
		}
		return rpcu, rpcpw, nil
	case PTPIVX:
		fmt.Println("Populating " + CPIVXConfFile + " for initial setup...")

		// Add rpcuser info if required
		b, err := StringExistsInFile(cRPCUserStr+"=", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcu = "pivxrpc"
			if err := WriteTextToFile(chd+CPIVXConfFile, cRPCUserStr+"="+rpcu); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcu, err = GetStringAfterStrFromFile(cRPCUserStr+"=", chd+CPIVXConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add rpcpassword info if required
		b, err = StringExistsInFile(cRPCPasswordStr+"=", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcpw = rand.String(20)
			if err := WriteTextToFile(chd+CPIVXConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, ""); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcpw, err = GetStringAfterStrFromFile(cRPCPasswordStr+"=", chd+CPIVXConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add daemon=1 info if required
		b, err = StringExistsInFile("daemon=1", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, "daemon=1"); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, ""); err != nil {
				log.Fatal(err)
			}
		}
		// Add server=1 info if required
		b, err = StringExistsInFile("server=1", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, "server=1"); err != nil {
				log.Fatal(err)
			}
		}
		// Add rpcallowip= info if required
		b, err = StringExistsInFile("rpcallowip=", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, "rpcallowip=192.168.1.0/255.255.255.0"); err != nil {
				log.Fatal(err)
			}
		}

		// Add rpcport= info if required
		b, err = StringExistsInFile("rpcport=", chd+CPIVXConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CPIVXConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CPIVXConfFile, "rpcport="+CPIVXRPCPort); err != nil {
				log.Fatal(err)
			}
		}
		return rpcu, rpcpw, nil
	case PTTrezarcoin:
		fmt.Println("Populating " + CTrezarcoinConfFile + " for initial setup...")

		// Add rpcuser info if required
		b, err := StringExistsInFile(cRPCUserStr+"=", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcu = "trezarcoinrpc"
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, cRPCUserStr+"="+rpcu); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcu, err = GetStringAfterStrFromFile(cRPCUserStr+"=", chd+CTrezarcoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add rpcpassword info if required
		b, err = StringExistsInFile(cRPCPasswordStr+"=", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			rpcpw = rand.String(20)
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, ""); err != nil {
				log.Fatal(err)
			}
		} else {
			rpcpw, err = GetStringAfterStrFromFile(cRPCPasswordStr+"=", chd+CTrezarcoinConfFile)
			if err != nil {
				return "", "", fmt.Errorf("unable to search for text in file - %v", err)
			}
		}

		// Add daemon=1 info if required
		b, err = StringExistsInFile("daemon=1", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, "daemon=1"); err != nil {
				log.Fatal(err)
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, ""); err != nil {
				log.Fatal(err)
			}
		}
		// Add server=1 info if required
		b, err = StringExistsInFile("server=1", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, "server=1"); err != nil {
				log.Fatal(err)
			}
		}
		// Add rpcallowip= info if required
		b, err = StringExistsInFile("rpcallowip=", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, "rpcallowip=192.168.1.0/255.255.255.0"); err != nil {
				log.Fatal(err)
			}
		}
		// Add rpcport= info if required
		b, err = StringExistsInFile("rpcport=", chd+CTrezarcoinConfFile)
		if err != nil {
			return "", "", fmt.Errorf("unable to search for text in file - %v", err)
		}
		if !b {
			if !bFileHasBeenBU {
				bFileHasBeenBU = true
				if err := BackupFile(chd, CTrezarcoinConfFile, false); err != nil {
					return "", "", fmt.Errorf("unable to backup file - %v", err)
				}
			}
			if err := WriteTextToFile(chd+CTrezarcoinConfFile, "rpcport="+CTrezarcoinRPCPort); err != nil {
				log.Fatal(err)
			}
		}
		return rpcu, rpcpw, nil
	default:
		err = errors.New("unable to determine ProjectType")
	}
	return "", "", nil
}

func AllProjectBinaryFilesExists() (bool, error) {
	//abf, err := GetAppsBinFolder()
	//if err != nil {
	//	return false, fmt.Errorf("Unable to GetAppsBinFolder - %v ", err)
	//}

	ex, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("Unable to retrieve running binary: %v ", err)
	}
	abf := AddTrailingSlash(filepath.Dir(ex))

	bwconf, err := GetConfigStruct("", false)
	if err != nil {
		return false, fmt.Errorf("unable to GetConfigStruct - %v", err)
	}
	switch bwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			if !FileExists(abf + CDiviCliFileWin) {
				return false, nil
			}
			if !FileExists(abf + CDiviDFileWin) {
				return false, nil
			}
			if !FileExists(abf + CDiviTxFileWin) {
				return false, nil
			}
		} else {
			if !FileExists(abf + CDiviCliFile) {
				return false, nil
			}
			if !FileExists(abf + CDiviDFile) {
				return false, nil
			}
			if !FileExists(abf + CDiviTxFile) {
				return false, nil
			}
		}
	case PTFeathercoin:
		if runtime.GOOS == "windows" {
			if !FileExists(abf + CFeathercoinCliFileWin) {
				return false, nil
			}
			if !FileExists(abf + CFeathercoinDFileWin) {
				return false, nil
			}
			if !FileExists(abf + CFeathercoinTxFileWin) {
				return false, nil
			}
		} else {
			if !FileExists(abf + CFeathercoinCliFile) {
				return false, nil
			}
			if !FileExists(abf + CFeathercoinDFile) {
				return false, nil
			}
			if !FileExists(abf + CFeathercoinTxFile) {
				return false, nil
			}
		}
	case PTPhore:
		if runtime.GOOS == "windows" {
			if !FileExists(abf + CPhoreCliFileWin) {
				return false, nil
			}
			if !FileExists(abf + CPhoreDFileWin) {
				return false, nil
			}
			if !FileExists(abf + CPhoreTxFileWin) {
				return false, nil
			}
		} else {
			if !FileExists(abf + CPhoreCliFile) {
				return false, nil
			}
			if !FileExists(abf + CPhoreDFile) {
				return false, nil
			}
			if !FileExists(abf + CPhoreTxFile) {
				return false, nil
			}
		}
	case PTPIVX:
		if runtime.GOOS == "windows" {
			if !FileExists(abf + CPIVXCliFileWin) {
				return false, nil
			}
			if !FileExists(abf + CPIVXDFileWin) {
				return false, nil
			}
			if !FileExists(abf + CPIVXTxFileWin) {
				return false, nil
			}
		} else {
			if !FileExists(abf + CPIVXCliFile) {
				return false, nil
			}
			if !FileExists(abf + CPIVXDFile) {
				return false, nil
			}
			if !FileExists(abf + CPIVXTxFile) {
				return false, nil
			}
		}
	case PTTrezarcoin:
		if runtime.GOOS == "windows" {
			if !FileExists(abf + CTrezarcoinCliFileWin) {
				return false, nil
			}
			if !FileExists(abf + CTrezarcoinDFileWin) {
				return false, nil
			}
			if !FileExists(abf + CTrezarcoinTxFileWin) {
				return false, nil
			}
		} else {
			if !FileExists(abf + CTrezarcoinCliFile) {
				return false, nil
			}
			if !FileExists(abf + CTrezarcoinDFile) {
				return false, nil
			}
			if !FileExists(abf + CTrezarcoinTxFile) {
				return false, nil
			}
		}
	default:
		err = errors.New("unable to determine ProjectType")
	}

	return true, nil
}

func updateAUDPriceInfo() error {
	resp, err := http.Get("https://api.exchangeratesapi.io/latest?base=USD&symbols=AUD")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &gPricePerCoinAUD)
	if err != nil {
		return err
	}
	return errors.New("unable to updateAUDPriceInfo")
}

func updateGBPPriceInfo() error {
	resp, err := http.Get("https://api.exchangeratesapi.io/latest?base=USD&symbols=GBP")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &gPricePerCoinGBP)
	if err != nil {
		return err
	}
	return errors.New("unable to updateGBPPriceInfo")
}

// WalletHardFix - Deletes the local blockchain and forces it to sync again, a re-index should be performed first
func WalletHardFix() error {
	// Stop divid if it's running
	if err := StopCoinDaemon(false); err != nil {
		return fmt.Errorf("unable to StopDiviD: %v", err)
	}

	chf, err := GetCoinHomeFolder(APPTCLI)
	if err != nil {
		return fmt.Errorf("unable to get coin home folder: %v", err)
	}

	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return err
	}
	switch gwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			//rm -r ~/.divi/blocks
			if err := os.RemoveAll(chf + "blocks"); err != nil {
				return fmt.Errorf("unable to remove the blocks folder: %v", err)
			}
			//rm -r ~/.divi/chainstate
			if err := os.RemoveAll(chf + "chainstate"); err != nil {
				return fmt.Errorf("unable to remove the chainstate folder: %v", err)
			}
			//rm -r ~/.divi/database
			if err := os.RemoveAll(chf + "database"); err != nil {
				return fmt.Errorf("unable to remove the database folder: %v", err)
			}
			//rm -r ~/.divi/sporks
			if err := os.RemoveAll(chf + "sporks"); err != nil {
				return fmt.Errorf("unable to remove the sporks folder: %v", err)
			}
			//rm -r ~/.divi/zerocoin
			if err := os.RemoveAll(chf + "zerocoin"); err != nil {
				return fmt.Errorf("unable to remove the zerocoin folder: %v", err)
			}

			//rm -r ~/.divi/db.log
			if err := os.Remove(chf + "db.log"); err != nil {
				return fmt.Errorf("unable to remove the db.log file: %v", err)
			}
			//rm -r ~/.divi/debug.log
			if err := os.Remove(chf + "debug.log"); err != nil {
				return fmt.Errorf("unable to remove the debug.log file: %v", err)
			}
			//rm -r ~/.divi/fee_estimates.dat
			if err := os.Remove(chf + "fee_estimates.dat"); err != nil {
				return fmt.Errorf("unable to remove the fee_estimates.dat file: %v", err)
			}
			//rm -r ~/.divi/peers.dat
			if err := os.Remove(chf + "peers.dat"); err != nil {
				return fmt.Errorf("unable to remove the peers.dat file: %v", err)
			}
			//rm -r ~/.divi/mncache.dat
			if err := os.Remove(chf + "mncache.dat"); err != nil {
				return fmt.Errorf("unable to remove the mncache.dat file: %v", err)
			}
			//rm -r ~/.divi/mnpayments.dat
			if err := os.Remove(chf + "mnpayments.dat"); err != nil {
				return fmt.Errorf("unable to remove the mnpayments.dat file: %v", err)
			}
			//rm -r ~/.divi/netfulfilled.dat
			if err := os.Remove(chf + "netfulfilled.dat"); err != nil {
				return fmt.Errorf("unable to remove the netfulfilled.dat file: %v", err)
			}

			// Now start the divid daemon again...
			os.Exit(0)
			//if err := RunCoinDaemon(false); err != nil {
			//	log.Fatalf("failed to run divid: %v", err)
			//}
		}
	case PTPhore:
	case PTPIVX:
	case PTTrezarcoin:
	default:
		err = errors.New("unable to determine ProjectType")
	}

	return nil
}

func WalletFix(wft WalletFixType) error {
	// Stop divid if it's running
	if err := StopCoinDaemon(false); err != nil {
		return fmt.Errorf("unable to StopDiviD: %v", err)
	}

	//abf, _ := GetAppsBinFolder()

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Unable to retrieve running binary: %v ", err)
	}
	abf := AddTrailingSlash(filepath.Dir(ex))

	coind, err := GetCoinDaemonFilename(APPTCLI)
	if err != nil {
		return fmt.Errorf("unable to GetCoinDaemonFilename - %v", err)
	}

	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return err
	}
	switch gwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			var arg1 string
			switch wft {
			case WFTReIndex:
				arg1 = "-reindex"
			case WFTReSync:
				arg1 = "-resync"
			}

			cRun := exec.Command(abf+coind, arg1)
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to run divid -reindex: %v", err)
			}
		}
	case PTFeathercoin:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			var arg1 string
			switch wft {
			case WFTReIndex:
				arg1 = "-reindex"
			case WFTReSync:
				arg1 = "-resync"
			}

			cRun := exec.Command(abf+coind, arg1)
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to run divid -reindex: %v", err)
			}
		}
	case PTPIVX:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			cRun := exec.Command(abf+coind, "-reindex")
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to run pivxd -reindex: %v", err)
			}
		}
	case PTTrezarcoin:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			cRun := exec.Command(abf+coind, "-reindex")
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to run trezardcoind -reindex: %v", err)
			}
		}
	default:
		err = errors.New("unable to determine ProjectType")
	}

	return nil
}

func runDCCommand(cmdBaseStr, cmdStr, waitingStr string, attempts int) (string, error) {
	var err error
	//var s string = waitingStr
	for i := 0; i < attempts; i++ {
		cmd := exec.Command(cmdBaseStr, cmdStr)
		out, err := cmd.CombinedOutput()

		if err == nil {
			return string(out), err
		}
		fmt.Printf("\r"+waitingStr+" %d/"+strconv.Itoa(attempts), i)

		time.Sleep(3 * time.Second)
	}

	return "", err
}

func runDCCommandWithValue(cmdBaseStr, cmdStr, valueStr, waitingStr string, attempts int) (string, error) {
	var err error
	//var s string = waitingStr
	for i := 0; i < attempts; i++ {
		cmd := exec.Command(cmdBaseStr, cmdStr, valueStr)
		out, err := cmd.CombinedOutput()

		if err == nil {
			return string(out), err
		}
		fmt.Printf("\r"+waitingStr+" %d/"+strconv.Itoa(attempts), i)
		time.Sleep(3 * time.Second)
	}

	return "", err
}

// RunCoinDaemon - Run the coins Daemon e.g. Run divid
func RunCoinDaemon(displayOutput bool) error {
	idr, _, _ := IsCoinDaemonRunning()
	if idr == true {
		// Already running...
		return nil
	}

	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return err
	}
	//abf, _ := GetAppsBinFolder()
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Unable to retrieve running binary: %v ", err)
	}
	abf := AddTrailingSlash(filepath.Dir(ex))

	switch gwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			//_ = exec.Command(GetAppsBinFolder() + cDiviDFileWin)
			fp := abf + CDiviDFileWin
			cmd := exec.Command("cmd.exe", "/C", "start", "/b", fp)
			if err := cmd.Run(); err != nil {
				return err
			}
		} else {
			if displayOutput {
				fmt.Println("Attempting to run the divid daemon...")
			}

			cmdRun := exec.Command(abf + CDiviDFile)
			stdout, err := cmdRun.StdoutPipe()
			if err != nil {
				return err
			}
			err = cmdRun.Start()
			if err != nil {
				return err
			}

			buf := bufio.NewReader(stdout) // Notice that this is not in a loop
			num := 1
			for {
				line, _, _ := buf.ReadLine()
				if num > 3 {
					os.Exit(0)
				}
				num++
				if string(line) == "DIVI server starting" {
					if displayOutput {
						fmt.Println("DIVI server starting")
					}
					return nil
				} else {
					return errors.New("unable to start Divi server: " + string(line))
				}
			}
		}
	case PTFeathercoin:
		if runtime.GOOS == "windows" {
			//_ = exec.Command(GetAppsBinFolder() + cDiviDFileWin)
			fp := abf + CFeathercoinDFileWin
			cmd := exec.Command("cmd.exe", "/C", "start", "/b", fp)
			if err := cmd.Run(); err != nil {
				return err
			}
		} else {
			if displayOutput {
				fmt.Println("Attempting to run the feathercoind daemon...")
			}

			cmdRun := exec.Command(abf + CFeathercoinDFile)
			stdout, err := cmdRun.StdoutPipe()
			if err != nil {
				return err
			}
			err = cmdRun.Start()
			if err != nil {
				return err
			}

			buf := bufio.NewReader(stdout) // Notice that this is not in a loop
			num := 1
			for {
				line, _, _ := buf.ReadLine()
				if num > 3 {
					os.Exit(0)
				}
				num++
				if string(line) == "Feathercoin server starting" {
					if displayOutput {
						fmt.Println("Feathercoin server starting")
					}
					return nil
				} else {
					return errors.New("unable to start Feathercoin server: " + string(line))
				}
			}
		}
	case PTPhore:
		if runtime.GOOS == "windows" {
			fp := abf + CDiviDFileWin
			cmd := exec.Command("cmd.exe", "/C", "start", "/b", fp)
			if err := cmd.Run(); err != nil {
				return err
			}
		} else {
			if displayOutput {
				fmt.Println("Attempting to run the phored daemon...")
			}

			cmdRun := exec.Command(abf + CPhoreDFile)
			stdout, err := cmdRun.StdoutPipe()
			if err != nil {
				return err
			}
			err = cmdRun.Start()
			if err != nil {
				return err
			}

			buf := bufio.NewReader(stdout) // Notice that this is not in a loop
			num := 1
			for {
				line, _, _ := buf.ReadLine()
				if num > 3 {
					os.Exit(0)
				}
				num++
				if string(line) == "Phore server starting" {
					if displayOutput {
						fmt.Println("Phore server starting")
					}
					return nil
				} else {
					return errors.New("unable to start Phore server: " + string(line))
				}
			}
		}
	case PTTrezarcoin:
		if runtime.GOOS == "windows" {
			fp := abf + CDiviDFileWin
			cmd := exec.Command("cmd.exe", "/C", "start", "/b", fp)
			if err := cmd.Run(); err != nil {
				return err
			}
		} else {
			if displayOutput {
				fmt.Println("Attempting to run the trezarcoin daemon...")
			}

			cmdRun := exec.Command(abf + CTrezarcoinDFile)
			stdout, err := cmdRun.StdoutPipe()
			if err != nil {
				return err
			}
			err = cmdRun.Start()
			if err != nil {
				return err
			}

			buf := bufio.NewReader(stdout) // Notice that this is not in a loop
			num := 1
			for {
				line, _, _ := buf.ReadLine()
				if num > 3 {
					os.Exit(0)
				}
				num++
				if string(line) == "Trezarcoin server starting" {
					if displayOutput {
						fmt.Println("Trezarcoin server starting")
					}
					return nil
				} else {
					return errors.New("unable to start Trezarcoin server: " + string(line))
				}
			}
		}
	default:
		err = errors.New("unable to determine ProjectType")
	}
	return nil
}

// stopCoinDaemon - Stops the coin daemon (e.g. divid) from running
func StopCoinDaemon(displayOutput bool) error {
	idr, _, _ := IsCoinDaemonRunning() //DiviDRunning()
	if idr != true {
		// Not running anyway ...
		return nil
	}

	//abf, _ := GetAppsBinFolder()
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Unable to retrieve running binary: %v ", err)
	}
	abf := AddTrailingSlash(filepath.Dir(ex))

	coind, err := GetCoinDaemonFilename(APPTCLI)
	if err != nil {
		return fmt.Errorf("unable to GetCoinDaemonFilename - %v", err)
	}

	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return err
	}
	switch gwconf.ProjectType {
	case PTDivi:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			for i := 0; i < 50; i++ {
				cRun := exec.Command(abf+CDiviCliFile, "stop")
				_ = cRun.Run()

				sr, _, _ := IsCoinDaemonRunning() //DiviDRunning()
				if !sr {
					// Lets wait a little longer before returning
					time.Sleep(3 * time.Second)
					return nil
				}
				if displayOutput {
					fmt.Printf("\rWaiting for divid server to stop %d/"+strconv.Itoa(50), i+1)
				}
				time.Sleep(3 * time.Second)
			}
		}
	case PTFeathercoin:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			for i := 0; i < 50; i++ {
				cRun := exec.Command(abf+CFeathercoinCliFile, "stop")
				_ = cRun.Run()

				sr, _, _ := IsCoinDaemonRunning()
				if !sr {
					// Lets wait a little longer before returning
					time.Sleep(3 * time.Second)
					return nil
				}
				if displayOutput {
					fmt.Printf("\rWaiting for feathercoind server to stop %d/"+strconv.Itoa(50), i+1)
				}
				time.Sleep(3 * time.Second)
			}
		}
	case PTPhore:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			for i := 0; i < 50; i++ {
				cRun := exec.Command(abf+CPhoreCliFile, "stop")
				_ = cRun.Run()

				sr, _, _ := IsCoinDaemonRunning()
				if !sr {
					// Lets wait a little longer before returning
					time.Sleep(3 * time.Second)
					return nil
				}
				if displayOutput {
					fmt.Printf("\rWaiting for phored server to stop %d/"+strconv.Itoa(50), i+1)
				}
				time.Sleep(3 * time.Second)
			}
		}
	case PTPIVX:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			cRun := exec.Command(abf+CPIVXCliFile, "stop")
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to StopPIVXD:%v", err)
			}

			for i := 0; i < 50; i++ {
				sr, _, _ := IsCoinDaemonRunning()
				if !sr {
					return nil
				}
				if displayOutput {
					fmt.Printf("\rWaiting for pivxd server to stop %d/"+strconv.Itoa(50), i+1)
				}
				time.Sleep(3 * time.Second)

			}
		}
	case PTTrezarcoin:
		if runtime.GOOS == "windows" {
			// TODO Complete for Windows
		} else {
			cRun := exec.Command(abf+CTrezarcoinCliFile, "stop")
			if err := cRun.Run(); err != nil {
				return fmt.Errorf("unable to StopCoinDaemon:%v", err)
			}

			for i := 0; i < 50; i++ {
				sr, _, _ := IsCoinDaemonRunning() //DiviDRunning()
				if !sr {
					return nil
				}
				if displayOutput {
					fmt.Printf("\rWaiting for "+coind+" server to stop %d/"+strconv.Itoa(50), i+1)
				}
				time.Sleep(3 * time.Second)

			}
		}
	default:
		err = errors.New("unable to determine ProjectType")
	}

	return nil
}

// RunInitialDaemon - Runs the divid Daemon for the first time to populate the divi.conf file
func RunInitialDaemon() (rpcuser, rpcpassword string, err error) {
	ex, err := os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("Unable to retrieve running binary: %v ", err)
	}
	abf := AddTrailingSlash(filepath.Dir(ex))

	coind, err := GetCoinDaemonFilename(APPTCLI)
	if err != nil {
		return "", "", fmt.Errorf("unable to GetCoinDaemonFilename - %v", err)
	}

	gwconf, err := GetConfigStruct("", false)
	if err != nil {
		return "", "", fmt.Errorf("unable to GetConfigStruct - %v", err)
	}
	switch gwconf.ProjectType {
	case PTDivi:
		//Run divid for the first time, so that we can get the outputted info to build the conf file
		fmt.Println("About to run " + coind + " for the first time...")
		cmdDividRun := exec.Command(abf + CDiviDFile)
		out, _ := cmdDividRun.CombinedOutput()
		fmt.Println("Populating " + CDiviConfFile + " for initial setup...")

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		var rpcuser, rpcpw string
		for scanner.Scan() {
			s := scanner.Text()
			if strings.Contains(s, cRPCUserStr) {
				rpcuser = strings.ReplaceAll(s, cRPCUserStr+"=", "")
			}
			if strings.Contains(s, cRPCPasswordStr) {
				rpcpw = strings.ReplaceAll(s, cRPCPasswordStr+"=", "")
			}
		}

		chd, _ := GetCoinHomeFolder(APPTCLI)

		if err := WriteTextToFile(chd+CDiviConfFile, cRPCUserStr+"="+rpcuser); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, cRPCPasswordStr+"="+rpcpw); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, ""); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, "daemon=1"); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, ""); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, "server=1"); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, "rpcallowip=0.0.0.0/0"); err != nil {
			log.Fatal(err)
		}
		if err := WriteTextToFile(chd+CDiviConfFile, "rpcport=8332"); err != nil {
			log.Fatal(err)
		}

		// Now get a list of the latest "addnodes" and add them to the file:
		// I've commented out the below, as I think it might cause future issues with blockchain syncing,
		// because, I think that the ipaddresses in the conf file are used before any others are picked up,
		// so, it's possible that they could all go, and then cause issues.

		// gdc.AddToLog(lfp, "Adding latest master nodes to "+gdc.CDiviConfFile)
		// addnodes, _ := gdc.GetAddNodes()
		// sAddnodes := string(addnodes[:])
		// gdc.WriteTextToFile(dhd+gdc.CDiviConfFile, sAddnodes)

		return rpcuser, rpcpw, nil
	case PTTrezarcoin:
		//Run divid for the first time, so that we can get the outputted info to build the conf file
		fmt.Println("Attempting to run " + coind + " for the first time...")
		cmdTrezarCDRun := exec.Command(abf + coind)
		if err := cmdTrezarCDRun.Start(); err != nil {
			return "", "", fmt.Errorf("failed to start %v: %v", coind, err)
		}

		return "", "", nil

	default:
		err = errors.New("unable to determine ProjectType")
	}
	return "", "", nil
}

// StopDaemon - Send a "stop" to the daemon, and returns the result
func StopDaemon(cliConf *ConfStruct) (GenericRespStruct, error) {
	var respStruct GenericRespStruct

	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"stop\",\"params\":[]}")
	req, err := http.NewRequest("POST", "http://"+cliConf.ServerIP+":"+cliConf.Port, body)
	if err != nil {
		return respStruct, err
	}
	req.SetBasicAuth(cliConf.RPCuser, cliConf.RPCpassword)
	req.Header.Set("Content-Type", "text/plain;")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return respStruct, err
	}
	defer resp.Body.Close()
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return respStruct, err
	}
	err = json.Unmarshal(bodyResp, &respStruct)
	if err != nil {
		return respStruct, err
	}
	return respStruct, nil
}

// UnlockWallet - Used by the server to unlock the wallet
//func UnlockWallet(pword string, attempts int, forStaking bool) (bool, error) {
//	var err error
//	var s string = "waiting for wallet."
//	dbf, _ := gwc.GetAppsBinFolder(gwc.APPTServer)
//	app := dbf + gwc.CDiviCliFile
//	arg1 := cCommandUnlockWalletFS
//	arg2 := pword
//	arg3 := "0"
//	arg4 := "true"
//	for i := 0; i < attempts; i++ {
//
//		var cmd *exec.Cmd
//		if forStaking {
//			cmd = exec.Command(app, arg1, arg2, arg3, arg4)
//		} else {
//			cmd = exec.Command(app, arg1, arg2, arg3)
//		}
//		//fmt.Println("cmd = " + dbf + cDiviCliFile + cCommandUnlockWalletFS + `"` + pword + `"` + "0")
//		out, err := cmd.CombinedOutput()
//
//		fmt.Println("string = " + string(out))
//		//fmt.Println("error = " + err.Error())
//
//		if err == nil {
//			return true, err
//		}
//
//		if strings.Contains(string(out), "The wallet passphrase entered was incorrect.") {
//			return false, err
//		}
//
//		if strings.Contains(string(out), "Loading block index....") {
//			//s = s + "."
//			//fmt.Println(s)
//			fmt.Printf("\r"+s+" %d/"+strconv.Itoa(attempts), i+1)
//
//			time.Sleep(3 * time.Second)
//
//		}
//
//	}
//
//	return false, err
//}
