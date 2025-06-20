package main

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
	Smartctl struct {
		Version     []int    `json:"version"`
		SvnRevision string   `json:"svn_revision"`
		Platform    string   `json:"platform_info"`
		BuildInfo   string   `json:"build_info"`
		Argv        []string `json:"argv"`
		ExitStatus  int      `json:"exit_status"`
	} `json:"smartctl"`

	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`

	ModelName       string `json:"model_name"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`

	NVMePCIVendor struct {
		ID          int `json:"id"`
		SubsystemID int `json:"subsystem_id"`
	} `json:"nvme_pci_vendor"`

	NVMeIEEEOUIIdentifier   int64 `json:"nvme_ieee_oui_identifier"`
	NVMETotalCapacity       int64 `json:"nvme_total_capacity"`
	NVMeUnallocatedCapacity int64 `json:"nvme_unallocated_capacity"`
	NVMeControllerID        int   `json:"nvme_controller_id"`

	NVMeVersion struct {
		String string `json:"string"`
		Value  int    `json:"value"`
	} `json:"nvme_version"`

	NVMeNumberOfNamespaces int `json:"nvme_number_of_namespaces"`

	NVMENamespaces []struct {
		ID   int `json:"id"`
		Size struct {
			Blocks int64 `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"size"`
		Capacity struct {
			Blocks int64 `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"capacity"`
		Utilization struct {
			Blocks int64 `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"utilization"`
		FormattedLBASize int `json:"formatted_lba_size"`
		EUI64            struct {
			OUI   int   `json:"oui"`
			ExtID int64 `json:"ext_id"`
		} `json:"eui64"`
	} `json:"nvme_namespaces"`

	UserCapacity struct {
		Blocks int64 `json:"blocks"`
		Bytes  int64 `json:"bytes"`
	} `json:"user_capacity"`

	LogicalBlockSize int `json:"logical_block_size"`

	LocalTime struct {
		TimeT   int64  `json:"time_t"`
		Asctime string `json:"asctime"`
	} `json:"local_time"`

	SmartStatus struct {
		Passed bool `json:"passed"`
		NVMe   struct {
			Value int `json:"value"`
		} `json:"nvme"`
	} `json:"smart_status"`

	NVMESMARTHealthInfo struct {
		CriticalWarning         int   `json:"critical_warning"`
		Temperature             int   `json:"temperature"`
		AvailableSpare          int   `json:"available_spare"`
		AvailableSpareThreshold int   `json:"available_spare_threshold"`
		PercentageUsed          int   `json:"percentage_used"`
		DataUnitsRead           int64 `json:"data_units_read"`
		DataUnitsWritten        int64 `json:"data_units_written"`
		HostReads               int64 `json:"host_reads"`
		HostWrites              int64 `json:"host_writes"`
		ControllerBusyTime      int64 `json:"controller_busy_time"`
		PowerCycles             int64 `json:"power_cycles"`
		PowerOnHours            int64 `json:"power_on_hours"`
		UnsafeShutdowns         int64 `json:"unsafe_shutdowns"`
		MediaErrors             int64 `json:"media_errors"`
		NumErrLogEntries        int64 `json:"num_err_log_entries"`
		WarningTempTime         int64 `json:"warning_temp_time"`
		CriticalCompTime        int64 `json:"critical_comp_time"`
		TemperatureSensors      []int `json:"temperature_sensors"`
	} `json:"nvme_smart_health_information_log"`

	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`

	PowerCycleCount int64 `json:"power_cycle_count"`

	PowerOnTime struct {
		Hours int64 `json:"hours"`
	} `json:"power_on_time"`
}

type SMARTInfo struct {
	Device string         `json:"device"`
	Output SmartctlOutput `json:"output"`
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
