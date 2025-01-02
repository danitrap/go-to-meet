package browser

import (
	"fmt"
	"log"
	"os/exec"
)

func OpenMeetLink(link string) {
	cmd := fmt.Sprintf("open '%s'", link)
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		log.Printf("Error opening meet link: %v", err)
	}
}
