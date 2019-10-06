#!/usr/bin/env bash

_body() {
        if [[ "${PAM_TYPE}" == "open_session" ]]; then
        cat <<EOM
User:   ${PAM_USER}
IP:     ${PAM_RHOST}
TTY:    ${PAM_SERVICE}/${PAM_TTY}
Date:   $(date -u)
Server: $(uname -a)
Status: ${PAM_TYPE}
EOM
	else
	        echo "${PAM_USER} from ${PAM_RHOST} has quit"
	fi
}

{
        printf '*%s*\t`%s`\n' "SSH activity detected on" "$(hostname -f)";
        printf '```\n';
        printf '%s\n' "$(_body)";
        printf '```\n';

} | /usr/local/bin/hermes

exit 0
