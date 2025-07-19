<p align="center"><a src="https://github.com/wizsk/mujamalat/releases/latest"><img src="./pub/fav.png" width="150"></a></p>

<h4 align="center">A libre Arabic Lexions Server/Web-app.</h4>

# Mujamalat

Arabic Lexicons.


## Install

Download the laset version from [https://github.com/wizsk/mujamalat/releases/latest](https://github.com/wizsk/mujamalat/releases/latest) for your os.

### linux

```bash
# linux
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_linux_$(uname -m).tar.gz"
tar xf "mujamalat_linux_$(uname -m).tar.gz"
sudo mv mujamalat /usr/local/bin/ # or mv mujamalat ~/.local/bin/
```


### Macos

```sh
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_macos_$(uname -m).tar.gz"
tar xf "mujamalat_macos_$(uname -m).tar.gz"
sudo mv mujadalat /usr/local/bin
```

or you can install it in your `~/.bin`

```sh
cd /tmp
wget "https://github.com/wizsk/mujamalat/releases/latest/download/mujamalat_macos_$(uname -m).tar.gz"
tar xf "mujamalat_macos_$(uname -m).tar.gz"
mkdir -p ~/.bin
mv mujamalat ~/.bin
echo 'export PATH="$PATH:$HOME/.bin"' >> ~/.bash_profile  # or ~/.zshrc for zsh users
source ~/.bash_profile  # or `source ~/.zshrc` for zsh users
```

### Windows

Open an `Administrator PowerShell` prompt and paste the following command

Go to Windows Search, type `PowerShell`, then right-click on the PowerShell app
in the search results or click the small arrow (>) next to it, and select Run as Administrator.


```ps1
irm "https://raw.githubusercontent.com/wizsk/mujamalat/refs/heads/main/install.ps1" | iex
```

