General:
- --recursive is default for some commands
- only diff format is git
- remove hg support
- remove source command (use git instead)
- --include option to many commands
- errors output to stderr, not stdout
- all paths printed with OS-specific path separator (except dump)
- --force now global
- --output now global
- diff includes scripts
- archive includes scripts
- encrypt -> encrypted in chattr

Config file:
- rename sourceVCS to git
- use gpg.recipient instead of gpgRecipient

Considering:
- remove Keyring support // FIXME: add warning