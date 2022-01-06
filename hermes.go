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
	defaultWebHook string
	WebHook        = os.Getenv("HERMES_WEBHOOK")
	alias          = os.Getenv("HERMES_ALIAS")
	emoji          = os.Getenv("HERMES_EMOJI")
	advanced       bool
	argsfirst      bool

	debug bool

	HttpClient = &http.Client{Timeout: 30 * time.Second}

	CLIFlags flag.FlagSet
)

func init() {
	if WebHook == "" && defaultWebHook != "" {
		WebHook = defaultWebHook
	}
	if debug {
		temple.DebugHTTPRequests = true
	}
	temple.StartTracking()
}

func (t Templates) SetFlags() {
	CLIFlags.StringVar(&WebHook, "W", defaultWebHook, "full webhook URL (env=WEBHOOK)")
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
			t = append(t, struct {
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
			t = append(t, struct {
				S      string
				IsFile bool
			}{value, true})
			return nil
		},
	)
}

type POST struct {
	Alias   string `json:"alias,omitempty"`
	Avatar  string `json:"avatar,omitempty"`
	Channel string `json:"channel,omitempty"`
	RoomID  string `json:"roomId,omitempty"`
	Text    string `json:"text"`
	Emoji   string `json:"emoji,omitempty"`
}

func (t Templates) Process(d Data) (*bytes.Buffer, error) {
	msg, err := t.work(d)
	if err != nil {
		return nil, err
	}
	if advanced {
		return msg, nil
	}
	bodyBuf := new(bytes.Buffer)
	tpl := POST{Text: msg.String()}
	if emoji != "" {
		tpl.Emoji = emoji
	}
	if alias != "" {
		tpl.Alias = alias
	}
	if err := json.NewEncoder(bodyBuf).Encode(&tpl); err != nil {
		return nil, err
	}
	return bodyBuf, nil
}

func Send(postData io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, WebHook, postData)
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
