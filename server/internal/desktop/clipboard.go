package desktop

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/xevent"
)

const (
	ClipboardTextPlainTarget = "UTF8_STRING"
	ClipboardTextHtmlTarget  = "text/html"
)

func (manager *DesktopManagerCtx) ClipboardGetText() (*types.ClipboardText, error) {
	text, err := manager.ClipboardGetBinary(ClipboardTextPlainTarget)
	if err != nil {
		return nil, err
	}

	// Rich text must not always be available, can fail silently.
	html, _ := manager.ClipboardGetBinary(ClipboardTextHtmlTarget)

	return &types.ClipboardText{
		Text: string(text),
		HTML: string(html),
	}, nil
}

func (manager *DesktopManagerCtx) ClipboardSetText(data types.ClipboardText) error {
	// TODO: Refactor.
	// Current implementation is unable to set multiple targets. HTML
	// is set, if available. Otherwise plain text.

	if data.HTML != "" {
		return manager.ClipboardSetBinary(ClipboardTextHtmlTarget, []byte(data.HTML))
	}

	return manager.ClipboardSetBinary(ClipboardTextPlainTarget, []byte(data.Text))
}

func (manager *DesktopManagerCtx) ClipboardGetBinary(mime string) ([]byte, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-out", "-target", mime)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		return nil, fmt.Errorf("%s", msg)
	}

	return stdout.Bytes(), nil
}

func (manager *DesktopManagerCtx) ClipboardSetBinary(mime string, data []byte) error {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-in", "-target", mime)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	// TODO: Refactor.
	// We need to wait until the data came to the clipboard.
	wait := make(chan struct{})
	xevent.Emmiter.Once("clipboard-updated", func(payload ...any) {
		wait <- struct{}{}
	})

	err = cmd.Start()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		return fmt.Errorf("%s", msg)
	}

	_, err = stdin.Write(data)
	if err != nil {
		return err
	}

	stdin.Close()

	// TODO: Refactor.
	// cmd.Wait()
	<-wait

	return nil
}

func (manager *DesktopManagerCtx) ClipboardGetTargets() ([]string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-out", "-target", "TARGETS")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		return nil, fmt.Errorf("%s", msg)
	}

	var response []string
	targets := strings.Split(stdout.String(), "\n")
	for _, target := range targets {
		if target == "" {
			continue
		}

		if !strings.Contains(target, "/") {
			continue
		}

		response = append(response, target)
	}

	return response, nil
}
