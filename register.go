package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
)

const torrefreshcmd = "/usr/bin/printf \"AUTHENTICATE \"YOURTORPASSWORD\"\r\nSIGNAL NEWNYM\r\n\" | /usr/bin/nc 127.0.0.1 9051"
const password = "YOURPASSWORDFORYOURCREATEDACCOUNTS" //Password for all users
const checkrecaptcha = `console.info(document.querySelector("#register-form > div.spacer > div > div > div > iframe").contentDocument);`
const query = `var recap = document.querySelector('#register-form > div.spacer > div > div > div > iframe').contentWindow;
recap.document.body.OuterHTML;`
const watson_username = "YOURWATSONAPIKEY"
const watson_password = "YOURWATSONPASSWORD"
const audio_file_path = "/your/path/to/save/audiofiles/in/audio.mp3"

func main() {
start:
	fmt.Println("Starting...")
	if DisableProxy() == -1 {
		time.Sleep(3 * time.Minute)
		goto start
	}

	//List of usernames you want to register
	infile, err := os.OpenFile("users.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		time.Sleep(3 * time.Minute)
		goto start
	}
	defer infile.Close()

	//Output file with a list of all successfully registered users
	outfile, err := os.OpenFile("registered.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		time.Sleep(3 * time.Minute)
		goto start
	}

	defer outfile.Close()
	scanner := bufio.NewScanner(infile)

	for scanner.Scan() {
		EnableProxy()
		RefreshTor()
		CleanUpAudioDir()
		var site, res string
		var err error

		// Create context
		ctxt, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create chrome instance
		c, err := chromedp.New(ctxt, chromedp.WithRunnerOptions(runner.Flag("disable-web-security", true)))
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		//Remove current username from users.txt
		line, err := PopLine(infile)
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}
		fmt.Println("Processing: " + string(line))

		//Perform registration process
		err = c.Run(ctxt, FillRegistrationForm(scanner.Text(), &site, &res))
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		//Recaptcha stuff starts here
		//Disable proxy first
		//Use mouse recording to click ReCaptcha checkbox. Sometimes it verifies and sometimes it asks you to solve a ReCaptcha.
		//Record the process as if you have to click the checkbox and then go to audio ReCaptcha and download it to your audio folder (assume the worst)
		//If the checkbox verfies, the worst that happens is you see your mouse moving and clicking a few places...no biggy
		DisableProxy()
		_, err = exec.Command("/bin/bash", "-c", "cnee --replay -f mouse/checkbox.xnr").Output()
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		//Do speech recognition with IBM Watson and get transcript back
		fmt.Println("Performing audio recognition...")
		transcript := AudioRecogntion()
		if transcript == "-1" {
			fmt.Println("Failed audio recognition!")
		}

		//Put transcript text into the clipboard copy buffer
		clipboard.WriteAll(transcript)
		//Paste and click Verify on ReCaptcha using recorde mouse movement
		_, err = exec.Command("/bin/bash", "-c", "cnee --replay -f mouse/verify.xnr").Output()
		if err != nil {
			fmt.Println("Failed to verify!")
		}
		time.Sleep(2 * time.Second)
		//Recaptcha stuff ends here
		fmt.Println(res)

		//Enable proxy because we're now going to hit the register button to complete the process
		if EnableProxy() == -1 {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}
		time.Sleep(2 * time.Second)

		//Register mouse recording
		_, err = exec.Command("/bin/bash", "-c", "cnee --replay -f mouse/register.xnr").Output()
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		time.Sleep(5 * time.Second)

		//Add to registered.txt
		if _, err = outfile.WriteString(string(line)); err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}
		fmt.Println("Registered: " + string(line))

		// Shutdown chrome
		err = c.Shutdown(ctxt)
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		// Wait for chrome to finish
		err = c.Wait()
		if err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}

		if _, err = outfile.WriteString(scanner.Text() + "\n"); err != nil {
			time.Sleep(3 * time.Minute)
			cancel()
			goto start
		}
	}

	if err := scanner.Err(); err != nil {
		time.Sleep(3 * time.Minute)
		goto start
	}
}

func FillRegistrationForm(username string, site, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(`https://www.reddit.com/login`),
		chromedp.WaitVisible(`#user_reg`, chromedp.ByID),
		chromedp.WaitVisible(`#passwd_reg`, chromedp.ByID),
		chromedp.WaitVisible(`#passwd2_reg`, chromedp.ByID),
		chromedp.WaitVisible(`#email_reg`, chromedp.ByID),
		chromedp.Sleep(time.Duration(8239-rand.Intn(100)) * time.Millisecond),
		chromedp.Click(`#user_reg`, chromedp.NodeVisible),
		chromedp.SetValue(`#user_reg`, username, chromedp.ByID),
		chromedp.Sleep(time.Duration(3739-rand.Intn(100)) * time.Millisecond),
		chromedp.Click(`#passwd_reg`, chromedp.NodeVisible),
		chromedp.WaitVisible(`#register-form > div:nth-child(3) > div > span.c-form-control-feedback.c-form-control-feedback-success`, chromedp.ByID),
		chromedp.SetValue(`#passwd_reg`, password, chromedp.ByID),
		chromedp.Sleep(time.Duration(3739-rand.Intn(100)) * time.Millisecond),
		chromedp.Click(`#passwd2_reg`, chromedp.NodeVisible),
		chromedp.WaitVisible(`#register-form > div:nth-child(4) > div.c-form-control-feedback-wrapper > span.c-form-control-feedback.c-form-control-feedback-success`, chromedp.ByID),
		chromedp.SetValue(`#passwd2_reg`, password, chromedp.ByID),
		chromedp.Sleep(time.Duration(6739-rand.Intn(100)) * time.Millisecond),
		chromedp.Click(`#email_reg`, chromedp.NodeVisible),
		chromedp.SetValue(`#email_reg`, username+"@gmail.com", chromedp.ByID),
	}
}

func AudioRecogntion() string {
	audio_file, err := ioutil.ReadFile(audio_file_path)
	if err != nil {
		return "-1"
	}

	req, err := http.NewRequest("POST", "https://stream.watsonplatform.net/speech-to-text/api/v1/recognize?model=en-US_NarrowbandModel", bytes.NewBuffer(audio_file))
	if err != nil {
		return "-1"
	}
	req.SetBasicAuth("YOURWATSONAPI", "YOURWATSONPASSWORD")
	req.Header.Set("Content-Type", "audio/mp3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "-1"
	}
	defer resp.Body.Close()
	received_JSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "-1"
	}
	var apiresults APIResults
	var phrases bytes.Buffer
	json.Unmarshal([]byte(received_JSON), &apiresults)
	for _, result := range apiresults.Results {
		for ai, _ := range result.Alternatives {
			phrases.WriteString(result.Alternatives[ai].Transcript)
		}
	}
	return phrases.String()
}

type APIResults struct {
	Results []struct {
		Alternatives []struct {
			Transcript string `json:"transcript"`
		} `json:"alternatives"`
		Final bool `json:"final"`
	} `json:"results"`
	ResultIndex int `json:"result_index"`
}

func CleanUpAudioDir() {
	directory := "/your/path/to/save/audiofiles/in/"
	dir_read, _ := os.Open(directory)
	dir_files, _ := dir_read.Readdir(0)

	// Loop over the directory's files.
	for index := range dir_files {
		file_here := dir_files[index]
		file_name := file_here.Name()
		full_path := directory + file_name

		if strings.HasPrefix(file_name, "audio") {
			// Remove the file.
			os.Remove(full_path)
			fmt.Println("Removed file:", full_path)
		}
	}
}

func RefreshTor() {
	_, err := exec.Command("/bin/bash", "-c", torrefreshcmd).Output()
	if err != nil {
		panic(err)
	}
}

func EnableProxy() int {
	_, err := exec.Command("/bin/bash", "-c", "gsettings set org.gnome.system.proxy mode 'manual'").Output()
	if err != nil {
		return -1
	}
	_, err = exec.Command("/bin/bash", "-c", "gsettings set org.gnome.system.proxy.socks port 9050").Output()
	if err != nil {
		return -1
	}
	_, err = exec.Command("/bin/bash", "-c", "gsettings set org.gnome.system.proxy.socks host 'localhost'").Output()
	if err != nil {
		return -1
	}
	return 0
}

func DisableProxy() int {
	_, err := exec.Command("/bin/bash", "-c", "gsettings set org.gnome.system.proxy mode 'none'").Output()
	if err != nil {
		return -1
	}
	return 0
}

func PopLine(f *os.File) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, fi.Size()))

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}
	line, err := buf.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	nw, err := io.Copy(f, buf)
	if err != nil {
		return nil, err
	}
	err = f.Truncate(nw)
	if err != nil {
		return nil, err
	}
	err = f.Sync()
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	return []byte(line), nil
}
