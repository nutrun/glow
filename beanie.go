package bean

import "exec"
import "strings"
import "fmt"

func GetUser() string {

    var user string = ""
    r, e := exec.Command("whoami").CombinedOutput()
    if e != nil {
        panic(e.String())
    }
    user = strings.Trim(string(r), "\n ")
    return user
}

func GetTubeName(user, tube string) string {
    if tube == "" {
        tube = "all"
    }
    return fmt.Sprintf("%s-%s", user, tube) 
}

func GetQueue(hp string, tube string) (*beanstalk.Tube, error) {
    if hp == "" {
        hp = "sug-chifjm03:11300"
    }
    conn, err := beanstalk.Dial(hp)
    if err != nil {
        return nil, err
    }
    return beanstalk.NewTube(conn, GetTubeName(GetUser(), tube))
}

func SubmitWorker(partition string, command string, tube string, rank int) {
    e := exec.Command("sbatch", "-p", partition, bean-worker, "-t", tube, 
}
