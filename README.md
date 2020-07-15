# An autocomplete wrapper for kubectl
## Installation
To install:  
`brew install cyberbliss/tap/kubectl-ac`

To upgrade:  
`brew upgrade cyberbliss/tap/kubectl-ac`

## Usage
A watch server needs to be running - this acts as a local cache of the Kube resources for the contexts you specify. The most efficient way of running the watch server is to start it first, e.g. via a bash function. Example of start and stop functions that could be defined in your .bashrc or .zshrc file: 
```
startkac(){
  if [ $# -lt 1 ]; then
    echo "Need to specify at least one K8s context"
  else
    # if kubectl-ac is already running then kill it first
    local _pid=$(pgrep kubectl-ac)
    if [ ! -z "$_pid" ]; then
      kill -s TERM $_pid
    fi
    kubectl-ac watch --syslog $@ &
  fi
}
stopkac(){
  local _pid=$(pgrep kubectl-ac)
  if [ -z "$_pid" ]; then
    echo "kubectl-ac not running"
  else
    kill -s TERM $_pid
    echo "kubectl-ac stopped"
  fi
}
```
To invoke the autocomplete function:
`kubectl ac <resource type>`, e.g. `kubectl ac po` for Pods. Use `kubectl ac --help` for more details.  
Note: if the watch service isn't running it will get automatically started by the above command.
## Development

### Releasing
goreleaser (https://goreleaser.com/) is used to generate new releases and push them to Github. It also generates the homebrew tap formula.
1. Make/test changes in a feature branch
2. PR and merge into master
3. Switch to master and git pull locally
4. `git tag -a vn.n.n` where n.n.n obeys semver standards
5. `git push origin vn.n.n`
6. `goreleaser --rm-dist`