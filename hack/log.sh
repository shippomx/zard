export TERM=ansi

function log::info() {
    echo "$(tput setab 7)$(tput setaf 2)$(tput bold) INFO $(tput sgr 0)" $@
}

function log::warn() {
    echo "$(tput setab 7)$(tput setaf 3)$(tput bold) WARN $(tput sgr 0)" $@
}

function log::error() {
    echo "$(tput setab 7)$(tput setaf 1)$(tput bold) ERROR $(tput sgr 0)" "$(tput setaf 1) $@ $(tput sgr 0)"
}
