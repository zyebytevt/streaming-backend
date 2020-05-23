package main

import (
	"image/color"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/christopher-dG/go-obs-websocket"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lornajane/streamdeck-tricks/actionhandlers"
	sdactionhandlers "github.com/magicmonkey/go-streamdeck/actionhandlers"
	"github.com/magicmonkey/go-streamdeck"
	buttons "github.com/magicmonkey/go-streamdeck/buttons"
	"github.com/spf13/viper"

	_ "github.com/godbus/dbus"
	"github.com/sqp/pulseaudio"
)

var mqtt_client mqtt.Client
var obs_client obsws.Client
var pulse *pulseaudio.Client

// InitButtons sets up initial button prompts
func InitButtons(sd *streamdeck.StreamDeck) {
	// Initialise MQTT to use the shelf light features
	mqtt_client = connectMQTT()

	// Initialise OBS to use OBS features (requires websockets plugin in OBS)
	obs_client = connectOBS()

	if obs_client.Connected() == true {
		obs_client.AddEventHandler("SwitchScenes", func(e obsws.Event) {
			// Make sure to assert the actual event type.
			log.Info().Msg("new scene: " + e.(obsws.SwitchScenesEvent).SceneName)
		})
	}

	// Get some Audio Setup
	pulse = getPulseConnection()

	// shelf lights
	abutton := buttons.NewColourButton(color.RGBA{255, 0, 255, 255})
	abutton.SetActionHandler(&actionhandlers.MQTTAction{Colour: color.RGBA{255, 0, 255, 255}, Client: mqtt_client})
	sd.AddButton(8, abutton)

	bbutton := buttons.NewColourButton(color.RGBA{0, 0, 255, 255})
	bbutton.SetActionHandler(&actionhandlers.MQTTAction{Colour: color.RGBA{0, 0, 255, 255}, Client: mqtt_client})
	sd.AddButton(9, bbutton)

	cbutton := buttons.NewColourButton(color.RGBA{255, 255, 0, 255})
	cbutton.SetActionHandler(&actionhandlers.MQTTAction{Colour: color.RGBA{255, 255, 0, 255}, Client: mqtt_client})
	sd.AddButton(10, cbutton)

	// OBS
	o1action := &actionhandlers.OBSSceneAction{Scene: "Camera", Client: obs_client}
	o1button := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/camera.png")
	o1button.SetActionHandler(o1action)
	sd.AddButton(24, o1button)

	o2button := buttons.NewImageFileButton(viper.GetString("buttons.images") + "/screen-and-cam.png")
	o2action := &actionhandlers.OBSSceneAction{Scene: "Screenshare", Client: obs_client}
	o2button.SetActionHandler(o2action)
	sd.AddButton(25, o2button)

	// Command
	eyesbutton := buttons.NewTextButton("Eyes")
	eyesaction := &sdactionhandlers.CustomAction{}
	eyesaction.SetHandler(func (btn streamdeck.Button) {
		cmd := exec.Command("xeyes")
		cmd.Start()
	})
	eyesbutton.SetActionHandler(eyesaction)
	sd.AddButton(7, eyesbutton)

	/*
		// example of multiple actions
		thisActionHandler := &sdactionhandlers.ChainedAction{}
		thisActionHandler.AddAction(&sdactionhandlers.TextPrintAction{Label: "Purple press"})
		thisActionHandler.AddAction(&sdactionhandlers.ColourChangeAction{NewColour: color.RGBA{255, 0, 0, 255}})
		multiActionButton := buttons.NewColourButton(color.RGBA{255, 0, 255, 255})
		multiActionButton.SetActionHandler(thisActionHandler)
		sd.AddButton(0, multiActionButton)
	*/

}

func connectMQTT() mqtt.Client {
	log.Debug().Msg("Connecting to MQTT...")
	opts := mqtt.NewClientOptions().AddBroker("tcp://10.1.0.1:1883").SetClientID("go-streamdeck")
	mqtt_client = mqtt.NewClient(opts)
	if conn_token := mqtt_client.Connect(); conn_token.Wait() && conn_token.Error() != nil {
		log.Warn().Err(conn_token.Error()).Msg("Cannot connect to MQTT")
	}
	return mqtt_client
}

func connectOBS() obsws.Client {
	log.Debug().Msg("Connecting to OBS...")
	log.Info().Msgf("%#v\n", viper.Get("obs.host"))
	obs_client = obsws.Client{Host: "localhost", Port: 4444}
	err := obs_client.Connect()
	if err != nil {
		log.Warn().Err(err).Msg("Cannot connect to OBS")
	}
	return obs_client
}

/*
// MyButtonPress reacts to a button being pressed
func MyButtonPress(btnIndex int, sd *streamdeck.Device, err error) {
	switch btnIndex {
	case 0:
		sources, _ := pulse.Core().ListPath("Sources")

		for _, src := range sources {
			dev := pulse.Device(src) // Only use the first sink for the test.
			var name string
			var muted bool
			dev.Get("Name", &name)
			dev.Get("Mute", &muted)
			fmt.Println(src, muted, name)

			dev.Set("Mute", true)
		}
	}
}
*/

type AppPulse struct {
	Client *pulseaudio.Client
}

func getPulseConnection() *pulseaudio.Client {
	isLoaded, e := pulseaudio.ModuleIsLoaded()
	testFatal(e, "test pulse dbus module is loaded")
	if !isLoaded {
		e = pulseaudio.LoadModule()
		testFatal(e, "load pulse dbus module")
	}

	// Connect to the pulseaudio dbus service.
	pulse, e := pulseaudio.New()
	testFatal(e, "connect to the pulse service")
	return pulse
}

func closePulseConnection(pulse *pulseaudio.Client) {
	//defer pulseaudio.UnloadModule()
	defer pulse.Close()
}

func testFatal(e error, msg string) {
	if e != nil {
		log.Warn().Err(e).Msg(msg)
	}
}

