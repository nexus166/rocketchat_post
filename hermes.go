package hermes

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/szampardi/xprint/temple"
)

type (
	Data = struct {
		Args  []string
		Stdin string
	} //
	Templates []struct {
		S      string
		IsFile bool
	}
)

var (
	defaultWebHook = ""
	webHook        = os.Getenv("HERMES_WEBHOOK")
	alias          = os.Getenv("HERMES_ALIAS")
	emoji          = os.Getenv("HERMES_EMOJI")
	advanced       bool
	debug          bool

	_templates Templates

	HttpClient = &http.Client{Timeout: 30 * time.Second}
	argsfirst  bool

	CLIFlags flag.FlagSet
)

func setFlags() {
	CLIFlags.StringVar(&webHook, "W", defaultWebHook, "full webhook URL (env=WEBHOOK)")
	CLIFlags.StringVar(&alias, "alias", "", "alias")
	CLIFlags.StringVar(&emoji, "e", ":satellite:", "emoji")
	CLIFlags.BoolVar(&argsfirst, "a", false, "render arguments (if any) before stdin (if any), instead of the opposite")
	CLIFlags.BoolVar(&advanced, "A", false, "use raw output as POST body")
	CLIFlags.BoolVar(&debug, "D", false, "debug init and template rendering activities")
	CLIFlags.BoolVar(&temple.EnableUnsafeFunctions, "u", unsafeMode(), fmt.Sprintf("allow evaluation of dangerous template functions (%v)", temple.FnMap.UnsafeFuncs()))
	CLIFlags.Func(
		"t",
		`template(s) (string). this flag can be specified more than once.
the last template specified in the commandline will be executed,
the others can be accessed with the "template" Action.
`,
		func(value string) error {
			_templates = append(_templates, struct {
				S      string
				IsFile bool
			}{value, false})
			return nil
		},
	)
	CLIFlags.Func(
		"f",
		`template(s) (files). this flag can be specified more than once.
the last template specified in the commandline will be executed,
the others can be accessed with the "template" Action.
`,
		func(value string) error {
			_, err := os.Stat(value)
			if err != nil {
				return err
			}
			_templates = append(_templates, struct {
				S      string
				IsFile bool
			}{value, true})
			return nil
		},
	)
}

func init() {
	setFlags()
	if webHook == "" && defaultWebHook != "" {
		webHook = defaultWebHook
	}
	if debug {
		temple.DebugHTTPRequests = true
	}
	temple.StartTracking()
}

type postTpl struct {
	Alias   string `json:"alias,omitempty"`
	Avatar  string `json:"avatar,omitempty"`
	Channel string `json:"channel,omitempty"`
	RoomID  string `json:"roomId,omitempty"`
	Text    string `json:"text"`
	Emoji   string `json:"emoji,omitempty"`
}

func Post(d Data) error {
	msg, err := process(d, _templates)
	if err != nil {
		return err
	}
	if advanced {
		if err := send(msg); err != nil {
			return err
		}
	} else {
		bodyBuf := new(bytes.Buffer)
		tpl := postTpl{Text: msg.String()}
		if emoji != "" {
			tpl.Emoji = emoji
		}
		if alias != "" {
			tpl.Alias = alias
		}
		if err := json.NewEncoder(bodyBuf).Encode(&tpl); err != nil {
			return err
		}
		if err := send(bodyBuf); err != nil {
			return err
		}
	}
	return nil
}

func send(postData io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, webHook, postData)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	defer HttpClient.CloseIdleConnections()
	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stderr, resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code in response: %d", resp.StatusCode)
	}
	return nil
}
