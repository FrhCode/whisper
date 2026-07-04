package recorder

import "testing"

func TestAudioDevicesGrouped(t *testing.T) {
	log := `[dshow @ 000001] "Integrated Camera" (video)
[dshow @ 000001] DirectShow audio devices (some may be both video and audio devices)
[dshow @ 000001]  "Microphone Array (Realtek(R) Audio)"
[dshow @ 000001]     Alternative name "@device_cm_{xxx}"
[dshow @ 000001]  "Headset Microphone"
[dshow @ 000001] DirectShow video devices (some may be both video and audio devices)
[dshow @ 000001]  "USB Camera"`
	got := AudioDevices(log)
	if len(got) != 2 || got[0] != "Microphone Array (Realtek(R) Audio)" || got[1] != "Headset Microphone" {
		t.Fatalf("bad devices: %#v", got)
	}
}

func TestAudioDevicesInline(t *testing.T) {
	log := `[in#0 @ 000001bc41b19400] "Integrated Camera" (video)
[in#0 @ 000001bc41b19400]   Alternative name "@device_pnp_x"
[in#0 @ 000001bc41b19400] "Microphone Array (AMD Audio Device)" (audio)
[in#0 @ 000001bc41b19400]   Alternative name "@device_cm_x"
Error opening input file dummy.`
	got := AudioDevices(log)
	if len(got) != 1 || got[0] != "Microphone Array (AMD Audio Device)" {
		t.Fatalf("bad devices: %#v", got)
	}
}
