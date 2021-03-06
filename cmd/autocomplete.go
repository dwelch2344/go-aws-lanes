package cmd

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var autoCompleteCmd = &cobra.Command{
	Use:     "completion [shell]",
	Short:   "Generate shell completion configuration",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"comp"},

	Run: func(cmd *cobra.Command, args []string) {
		var (
			shell = path.Base(os.Getenv("SHELL"))

			fl    = cmd.Flags()
			nargs = fl.NArg()
		)

		if nargs >= 1 {
			shell = fl.Arg(0)
		}

		switch strings.ToLower(shell) {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			runCompletionZsh(cmd, os.Stdout)
		default:
			cmd.Printf("Unsupported shell: %s\n", shell)
		}
	},
}

// This doesn't seem to work for me. I snagged it from:
// https://github.com/kubernetes/kubernetes/blob/0153f8de1a17668cb8e0a38b02c81d77711a8285/pkg/kubectl/cmd/completion.go#L140
func runCompletionZsh(cmd *cobra.Command, out io.Writer) {
	out.Write([]byte("#compdef lanes\n"))

	zsh_initialization := `
__lanes_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}

__lanes_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__lanes_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}

__lanes_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}

__lanes_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__lanes_ltrim_colon_completions() {
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}

__lanes_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}

__lanes_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__lanes_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}

__lanes_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
		printf %q "$1"
	fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi

__lanes_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__lanes_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__lanes_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__lanes_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__lanes_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__lanes_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__lanes_type/g" \
	<<BASH_COMPLETION_EOF
`
	out.Write([]byte(zsh_initialization))

	cmd.GenBashCompletion(out)

	out.Write([]byte(`
BASH_COMPLETION_EOF
}

__lanes_bash_source <(__lanes_convert_bash_to_zsh)
_complete lanes 2>/dev/null
`))
}
