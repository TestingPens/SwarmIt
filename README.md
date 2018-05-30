# SwarmIt - A Google ReCaptcha Bypass Script
SwarmIt is a Golang project I created to learn the language and challenge myself to bypass Google's infamous ReCaptcha. The usecase for this project was to find a practical way to automate the process of registering a large number of social media accounts while bypassing ReCaptcha.
## Overview
Essentially, this was from a 24 hour hackathon I did fueled by too much caffeine. The script is written in Golang and uses the Chromedp browser automation library to simulate a human registering on Reddit.com. It uses Tor to cycle IP addresses and IBM's Watson project to perform audio recognition on ReCaptchas. Cnee was used to record and play mouse movements.
### Why Chromedp?
I know Python Selenium and wanted to try something new. Chromedp has some limitations such as being able to change frame perspectives but there are ways around that (Cnee). Chromedp is the easiest browser automation library I've ever used.
### Why Tor?
Most social media sites will limit the number of times you can register from a specific IP address. Tor is used with its control port open so that a new IP address can be obtained for each account registration. Traffic is proxied through Tor only for the registration process and then disabled prior to bypassing ReCaptcha. The reason for this decision was because Tor IP addresses have a very bad reputation score, affecting the difficulty of ReCaptchas you get. Thus, the bypass process is done without Tor.
### Why IBM Watson Audio Recognition?
I struggled for many hours with all of the praised image recognition services out there. The bottom line is that ReCaptcha image quality is far too poor to be able to accurately perform image recognition. It turns out that audio recognition is a lot easier to perform that image recognition. IBM Watson Audio recognition service produced the highest success rate.
### Why Cnee?
I tried simulating human movement of the mouse and clicking by altering acceleration, velocity and path curve but Google is just too smart at identifying automated behaviour. It may be the least elegant solution, however recording your mouse moving and clicking buttons gets the job done. I feel this kind of stuff isn't worth burning brain power on.
## My Setup to Make This Work
- Ubuntu with Gnome (required for enabling/disabling proxy settings)
- Tor with control port enabled (for recycling IP addresses without restarting Tor)
- Golang installed with Chromedp lib (https://github.com/chromedp/chromedp)
- Cnee installed
- Record mouse movements with Cnee for clicking buttons such as registration and on ReCaptcha
- IBM Watson API key and password
- A text file containing a large number of unique usernames (I used rsmangler to generate thousands)
- Lots of patience

I'm sure there are far better ways of doing this so please feel free to contribute.
