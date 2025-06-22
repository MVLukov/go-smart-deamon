package gosmartdeamon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type BlockDevice struct {
	Name       string        `json:"name"`
	KName      string        `json:"kname"`
	MajMin     string        `json:"maj:min"`
	RM         bool          `json:"rm"`
	Size       string        `json:"size"`
	Type       string        `json:"type"`
	Mountpoint string        `json:"mountpoint"`
	Children   []BlockDevice `json:"children,omitempty"`
}

type LSBLKOutput struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type SmartctlOutput struct {
	JsonFormatVersion []int    `json:"json_format_version"`
	Smartctl          Smartctl `json:"smartctl"`
	Device            Device   `json:"device"`

	ModelName        string   `json:"model_name"`
	SerialNumber     string   `json:"serial_number"`
	FirmwareVersion  string   `json:"firmware_version"`
	UserCapacity     Capacity `json:"user_capacity"`
	LogicalBlockSize int      `json:"logical_block_size,omitempty"`

	Temperature     *TemperatureWrapper `json:"temperature,omitempty"`
	PowerCycleCount int                 `json:"power_cycle_count,omitempty"`
	PowerOnTime     PowerOnTime         `json:"power_on_time,omitempty"`
	LocalTime       LocalTime           `json:"local_time"`

	SmartStatus *SmartStatus `json:"smart_status,omitempty"`

	// NVMe-specific
	NVMePCI             *NVMePCIVendor              `json:"nvme_pci_vendor,omitempty"`
	NVMeIEEE            *int                        `json:"nvme_ieee_oui_identifier,omitempty"`
	NVMeTotalCapacity   *int64                      `json:"nvme_total_capacity,omitempty"`
	NVMeUnallocCapacity *int64                      `json:"nvme_unallocated_capacity,omitempty"`
	NVMeControllerID    *int                        `json:"nvme_controller_id,omitempty"`
	NVMeVersion         *NVMeVersion                `json:"nvme_version,omitempty"`
	NVMeNamespaces      []NVMeNamespace             `json:"nvme_namespaces,omitempty"`
	NVMESmartLog        *NVMESmartHealthInformation `json:"nvme_smart_health_information_log,omitempty"`

	// SATA-specific
	ATAAttributes      *ATAAttributes `json:"ata_smart_attributes,omitempty"`
	ATASmartStatus     *ATAStatus     `json:"ata_smart_status,omitempty"`
	ATAErrorLogVersion *int           `json:"ata_error_count,omitempty"`
}

type Smartctl struct {
	Version     []int    `json:"version"`
	SVNRevision string   `json:"svn_revision"`
	Platform    string   `json:"platform_info"`
	BuildInfo   string   `json:"build_info"`
	Argv        []string `json:"argv"`
	ExitStatus  int      `json:"exit_status"`
}

type Device struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

type Capacity struct {
	Blocks int64 `json:"blocks"`
	Bytes  int64 `json:"bytes"`
}

type TemperatureWrapper struct {
	Current int `json:"current"`
}

type PowerOnTime struct {
	Hours int `json:"hours"`
}

type LocalTime struct {
	TimeT   int64  `json:"time_t"`
	Asctime string `json:"asctime"`
}

type SmartStatus struct {
	Passed bool `json:"passed"`
	NVMe   *struct {
		Value int `json:"value"`
	} `json:"nvme,omitempty"`
}

type NVMePCIVendor struct {
	ID          int `json:"id"`
	SubsystemID int `json:"subsystem_id"`
}

type NVMeVersion struct {
	String string `json:"string"`
	Value  int    `json:"value"`
}

type NVMeNamespace struct {
	ID               int      `json:"id"`
	Size             Capacity `json:"size"`
	Capacity         Capacity `json:"capacity"`
	Utilization      Capacity `json:"utilization"`
	FormattedLBASize int      `json:"formatted_lba_size"`
	EUI64            struct {
		Oui   int   `json:"oui"`
		ExtID int64 `json:"ext_id"`
	} `json:"eui64"`
}

type NVMESmartHealthInformation struct {
	CriticalWarning      int   `json:"critical_warning"`
	Temperature          int   `json:"temperature"`
	AvailableSpare       int   `json:"available_spare"`
	AvailableSpareThresh int   `json:"available_spare_threshold"`
	PercentageUsed       int   `json:"percentage_used"`
	DataUnitsRead        int64 `json:"data_units_read"`
	DataUnitsWritten     int64 `json:"data_units_written"`
	HostReads            int64 `json:"host_reads"`
	HostWrites           int64 `json:"host_writes"`
	ControllerBusyTime   int64 `json:"controller_busy_time"`
	PowerCycles          int   `json:"power_cycles"`
	PowerOnHours         int   `json:"power_on_hours"`
	UnsafeShutdowns      int   `json:"unsafe_shutdowns"`
	MediaErrors          int   `json:"media_errors"`
	NumErrLogEntries     int64 `json:"num_err_log_entries"`
	WarningTempTime      int   `json:"warning_temp_time"`
	CriticalCompTime     int   `json:"critical_comp_time"`
	TemperatureSensors   []int `json:"temperature_sensors"`
}

type ATAStatus struct {
	Passed bool `json:"passed"`
}

type ATAAttributes struct {
	Revision int `json:"revision"`
	Table    []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Value      int    `json:"value"`
		Worst      int    `json:"worst"`
		Thresh     int    `json:"thresh"`
		WhenFailed string `json:"when_failed"`
		Flags      struct {
			Value         int    `json:"value"`
			String        string `json:"string"`
			Prefailure    bool   `json:"prefailure"`
			UpdatedOnline bool   `json:"updated_online"`
			Performance   bool   `json:"performance"`
			ErrorRate     bool   `json:"error_rate"`
			EventCount    bool   `json:"event_count"`
			AutoKeep      bool   `json:"auto_keep"`
		} `json:"flags"`
		Raw struct {
			Value  int    `json:"value"`
			String string `json:"string"`
		} `json:"raw"`
	} `json:"table"`
}

type SMARTInfo struct {
	Device string
	Output SmartctlOutput
}

func getDevices() (LSBLKOutput, error) {
	cmd := exec.Command("lsblk", "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return LSBLKOutput{}, fmt.Errorf("smartctl error: %v", err)
	}

	result := LSBLKOutput{}
	json.Unmarshal(out, &result)

	return result, nil
}

func getSMART(device string) (SmartctlOutput, error) {
	cmd := exec.Command("smartctl", "--json", "-a", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return SmartctlOutput{}, fmt.Errorf("smartctl error: %v", err)
	}

	var result SmartctlOutput
	err = json.Unmarshal(out, &result)
	if err != nil {
		log.Fatal(err)
	}

	return result, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	smartInfo := []SMARTInfo{}

	devices, err := getDevices()
	if err != nil {
		panic(err)
	}

	for _, d := range devices.BlockDevices {
		if d.Name != "" {
			dPath := fmt.Sprintf("/dev/%s", d.Name)

			smart, _ := getSMART(dPath)
			if err != nil {
				log.Fatal(err)
				continue
			}

			smartInfo = append(smartInfo, SMARTInfo{
				Device: dPath,
				Output: smart,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(smartInfo)
}

func main() {

	http.HandleFunc("/smart", handler)
	fmt.Println("Running smartd-agent on :9090")
	http.ListenAndServe("localhost:9090", nil)
}
