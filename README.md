<p align="center"><a src="https://github.com/wizsk/mujamalat/releases/latest"><img src="./pub/fav.png" width="150"></a></p>

<h4 align="center">A libre Arabic Lexions Server/Web-app.</h4>

# Mujamalat

It includes,

- 6 Arabic lexicons: 
    - معجم الغني، معجم اللغة العربية المعاصرة، معجم الوسيط، معجم المحيط، مختار الصحاح، لسان العرب
- 2 English lexicon:
    - Lane Lexicon, Hanswehr
- 1 direct Arabic to English dictionary

You can try it out here. [https://mujamalat.onrender.com](https://mujamalat.onrender.com)

It's running on the hobby plan so, it may take some time to laod, and it may say it's booting up.


## Install

Download the laset version from [https://github.com/wizsk/mujamalat/releases/latest](https://github.com/wizsk/mujamalat/releases/latest) for your os.

### linux

```bash
# linux
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_linux_x86_64.tar.gz"
tar xf "mujamalat_linux_x86_64.tar.gz"
sudo mv mujamalat /usr/local/bin/ # or mv mujamalat ~/.local/bin/
```


### Macos

**For x86_64 intel**

```sh
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_macos_x86_64.tar.gz"
tar xf "mujamalat_macos_x86_64.tar.gz"
```
**For ARM**

```sh
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_macos_arm64.tar.gz"
tar xf "mujamalat_macos_arm64.tar.gz"
```
Then run these to move binnary to bin and add to the bin PATH.

```sh
mkdir -p ~/.bin
mv mujamalat ~/.bin
echo 'export PATH="$PATH:$HOME/.bin"' >> ~/.bash_profile  # or ~/.zshrc for zsh users
source ~/.bash_profile  # or `source ~/.zshrc` for zsh users
```

### Windows

Open an `Administrator PowerShell` prompt and paste the following command

Go to Windows Search, type `PowerShell`, then right-click on the PowerShell app
in the search results or click the small arrow (>) next to it, and select Run as Administrator.

**This will only work for x86_64.**

```ps1
irm "https://raw.githubusercontent.com/wizsk/mujamalat/refs/heads/main/install.ps1" | iex
```
