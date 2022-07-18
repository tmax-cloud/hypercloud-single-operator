package util

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/gomail.v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Clientset    *kubernetes.Clientset
	config       *restclient.Config
	ResourceList []string

	Trial_DueDate       int
	Trial_AlarmDateDiff int
	Trial_MailDate      int
)

func init() {
	var err error
	config, err = restclient.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	config.Burst = 100
	config.QPS = 100
	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	Trial_DueDate = getenv("TRIAL_DUEDATE", 30)
	Trial_AlarmDateDiff = getenv("TRIAL_ALARMDATEDIFF", 7)
	Trial_MailDate = Trial_DueDate - Trial_AlarmDateDiff
}

func getenv(key string, fallback int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	v, _ := strconv.Atoi(value)
	return v
}

const (
	TEST = "<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"    <meta charset=\"UTF-8\">\r\n" +
		"    <title>HyperCloud 서비스 신청 승인 완료</title>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<div style=\"border: #c5c5c8 0.06rem solid; border-bottom: 0; width: 42.5rem; height: 53.82rem; padding: 0 1.25rem\">\r\n" +
		"    <header>\r\n" +
		"        <div style=\"margin: 0;\">\r\n" +
		"            <p style=\"font-size: 1rem; font-weight: bold; color: #333333; line-height: 3rem; letter-spacing: 0; border-bottom: #c5c5c8 0.06rem solid;\">\r\n" +
		"                HyperCloud 서비스 신청 승인 완료\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </header>\r\n" +
		"    <section>\r\n" +
		"        <figure style=\"text-align: center;\">\r\n" +
		"            <img style=\"margin: 0.94rem 0;\"\r\n" +
		"                 src=\"cid:trial-approval\">\r\n" +
		"        </figure>\r\n" +
		"        <div style=\"width: 35.70rem; margin: 0 2.75rem;\">\r\n" +
		"            <p style=\"font-size: 1.5rem; font-weight: bold; line-height: 3rem;\">\r\n" +
		"                축하합니다.\r\n" +
		"            </p>\r\n" +
		"            <p style=\"line-height: 1.38rem;\">\r\n" +
		"                고객님의 Trial 서비스 신청이 성공적으로 승인되었습니다. <br>\r\n" +
		"                지금 바로 티맥스의 소프트웨어와 검증을 거친 오픈소스 서비스를 결합한 클라우드 플랫폼, <br>\r\n" +
		"                HyperCloud를 이용해 보세요. <br>\r\n" +
		"                <br>\r\n" +
		"                네임스페이스 이름 : <span style=\"font-weight: 600;\">%%NAMESPACE_NAME%%</span> <br>\r\n" +
		"                Trial 기한 : %%TRIAL_START_TIME%% ~ %%TRIAL_END_TIME%% <br>\r\n" +
		"                <br>\r\n" +
		"                리소스 정보 <br>\r\n" +
		"                -CPU : 1 Core <br>\r\n" +
		"                -Memory : 4 GIB <br>\r\n" +
		"                -Storage : 4 GIB <br>\r\n" +
		"                <br>\r\n" +
		"<!--                <span style=\"font-weight: 600;\">승인사유</span> <br>-->\r\n" +
		"                <br>\r\n" +
		"\r\n" +
		"                감사합니다. <br>\r\n" +
		"                TmaxCloud 드림.\r\n" +
		"            </p>\r\n" +
		"            <p style=\"margin: 3rem 0;\">\r\n" +
		"                <a href=\"https://console.tmaxcloud.com\">Tmax Console 바로가기 ></a>\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </section>\r\n" +
		"</div>\r\n" +
		"<footer style=\"background-color: #3669B3; width: 45.12rem; height: 1.88rem; font-size: 0.75rem; color: #FFFFFF; display: flex;\r\n" +
		"    align-items: center; justify-content: center;\">\r\n" +
		"    <div>\r\n" +
		"        COPYRIGHT2020. TMAX A&C., LTD. ALL RIGHTS RESERVED\r\n" +
		"    </div>\r\n" +
		"</footer>\r\n" +
		"</body>\r\n" +
		"</html>"
)

func SetTrialNSTimer(ns *v1.Namespace, client client.Client, reqLogger logr.Logger) {
	reqLogger.V(3).Info(" [Trial Timer] TrialNSTimer for Trial NS[ " + ns.Name + " ] Set Service Start")
	currentTime := time.Now()
	createTime := ns.CreationTimestamp.Time
	reqLogger.V(3).Info(" [Trial Timer] CreateTime of Trial NS[ " + ns.Name + " ] : " + createTime.String())
	mailTime := createTime.AddDate(0, 0, Trial_MailDate)
	// mailTime := createTime.Add(time.Second * 23)  // for test
	deleteTime := createTime.AddDate(0, 0, Trial_DueDate)
	// deleteTime := createTime.Add(time.Second * 30)  // for test

	if ns.Labels["period"] != "" {
		period, _ := strconv.Atoi(ns.Labels["period"])
		deleteTime = createTime.AddDate(0, 0, period*30)
		mailTime = deleteTime.AddDate(0, 0, -Trial_AlarmDateDiff)
	}
	if mailTime.After(currentTime) {
		time.AfterFunc(time.Duration((mailTime.UnixNano() - currentTime.UnixNano())), func() {
			reqLogger.V(3).Info(" [Trial Timer] Trial NameSpace [ " + ns.Name + " ] Mail Service before 1 weeks of deletion Start")
			nsFound := &v1.Namespace{}
			if err := client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, nsFound); err != nil && errors.IsNotFound(err) {
				reqLogger.V(3).Info(" [Trial Timer]  NameSpace [ " + ns.Name + " ] has Deleted, Nothing to do")
			}

			currentTimeinTimer := time.Now()
			periodinTimer, _ := strconv.Atoi(nsFound.Labels["period"])
			reqLogger.V(3).Info("[Trial Timer] currentTime in MailTimer of Trial NS[ " + ns.Name + " ] : " + currentTimeinTimer.String())
			reqLogger.V(3).Info("[Trial Timer] Period in MailTimer of Trial NS[ " + ns.Name + " ] : " + nsFound.Labels["period"])
			expectedCreationDate := currentTimeinTimer.AddDate(0, -periodinTimer, 7)

			if (expectedCreationDate.Second() - currentTimeinTimer.Second()) < 1000 {
				if nsFound.Labels != nil && nsFound.Labels["trial"] != "" && nsFound.Annotations != nil && nsFound.Annotations["owner"] != "" {
					reqLogger.V(3).Info(" [Trial Timer] Still Trial NameSpace, Send Info Mail to User [ " + nsFound.Annotations["owner"] + " ]")
					subject := " 신청해주신 Trial NameSpace [ " + nsFound.Name + " ] 만료 안내 "
					body := TRIAL_TIME_OUT_CONTENTS
					body = strings.ReplaceAll(body, "%%TRIAL_END_TIME%%", deleteTime.Format("2006-01-02"))
					SendMail(nsFound.Annotations["owner"], subject, body, "/img/service-timeout.png", "service-timeout", reqLogger)
				} else {
					reqLogger.V(3).Info(" [Trial Timer] Paid NameSpace, Nothing to do")
				}
			} else {
				reqLogger.V(3).Info(" [Trial Timer] Mail Timer wake up at the wrong date, Nothing to do")
			}
		})
		reqLogger.V(3).Info(" [Trial Timer] Set Trial NameSpace Sending Mail Timer Success ")
		reqLogger.V(3).Info(" [Trial Timer] MailSendTime for Trial NS[ " + ns.Name + " ] : " + mailTime.String())

		ns.Labels["mailSendDate"] = mailTime.Format("2006-01-02")
	} else {
		reqLogger.V(3).Info(" [Trial Timer] Mail for Alert Deletion for This Trial Namespace [" + ns.Name + "] already Sent to " + ns.Annotations["owner"])
	}

	if deleteTime.After(currentTime) {
		time.AfterFunc(time.Duration((deleteTime.UnixNano() - currentTime.UnixNano())), func() {
			reqLogger.V(3).Info(" [Trial Timer] Trial NameSpace [ " + ns.Name + " ] deletion Start")
			nsFound := &v1.Namespace{}
			if err := client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, nsFound); err != nil && errors.IsNotFound(err) {
				reqLogger.V(1).Error(err, " [Trial Timer]  NameSpace [ "+ns.Name+" ] has Deleted, Nothing to do")
			}

			currentTimeinTimer := time.Now()
			periodinTimer, _ := strconv.Atoi(nsFound.Labels["period"])
			reqLogger.V(3).Info("[Trial Timer] currentTime in DeleteTimer of Trial NS[ " + ns.Name + " ] : " + currentTimeinTimer.String())
			reqLogger.V(3).Info("[Trial Timer] Period in DeleteTimer of Trial NS[ " + ns.Name + " ] : " + nsFound.Labels["period"])
			expectedCreationDate := currentTimeinTimer.AddDate(0, -periodinTimer, 0)
			if (expectedCreationDate.Second() - currentTimeinTimer.Second()) < 1000 {
				if nsFound.Labels != nil && nsFound.Labels["trial"] != "" && nsFound.Annotations != nil && nsFound.Annotations["owner"] != "" {
					reqLogger.V(3).Info(" [Trial Timer] Still Trial NameSpace, Delete Expired Namespace [ " + nsFound.Name + " ]")
					if err := client.Delete(context.TODO(), nsFound); err != nil {
						reqLogger.V(1).Error(err, " [Trial Timer] Failed to Delete NameSpace [ "+ns.Name+" ]")
						panic(err)
					} else if err := client.Delete(context.TODO(), &rbacv1.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "CRB-" + nsFound.Name,
						},
					}); err != nil {
						reqLogger.V(1).Error(err, " [Trial Timer] Failed to Delete ClusterRoleBinding [ "+"CRB-"+nsFound.Name+" ]")
						panic(err)
					} else {
						reqLogger.V(3).Info(" [Trial Timer] Delete Expired Namespace [ " + nsFound.Name + " ] Success")
					}
				} else {
					reqLogger.V(3).Info(" [Trial Timer] Paid NameSpace, Nothing to do")
				}
			} else {
				reqLogger.V(3).Info(" [Trial Timer] Delete Timer wake up at the wrong date, Nothing to do")
			}
		})
		reqLogger.V(3).Info(" [Trial Timer] Set Trial NameSpace delete Timer Success ")
		reqLogger.V(3).Info(" [Trial Timer] DeletionTime for Trial NS[ " + ns.Name + " ] : " + deleteTime.String())

		ns.Labels["deletionDate"] = deleteTime.Format("2006-01-02")
		if err := client.Update(context.TODO(), ns); err != nil {
			reqLogger.V(1).Error(err, "[Trial Timer] Replace NameSpace for new Label Failed")
			panic(err)
		} else {
			reqLogger.V(3).Info(" [Trial Timer] Replace NameSpace for new Label Success ")
		}
	} else {
		reqLogger.V(3).Info(" [Trial Timer] This Trial Namespace [" + ns.Name + "] has Already Expired, Check Why This NameSpace is Still Exists")
	}
}

func SendMail(recipient string, subject string, body string, imgPath string, imgCid string, reqLogger logr.Logger) {
	reqLogger.V(3).Info(" Send Mail to User [ " + recipient + " ] Start")
	host := "mail.tmax.co.kr"
	port := 25
	sender := "no-reply-tc@tmax.co.kr"
	pw := "!@tcdnsdudxla11"

	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply-tc@tmax.co.kr")
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	// m.SetHeader("Content-Type", "image/svg")
	// m.SetHeader("Content-ID", "<"+imgCid+">")
	m.SetBody("text/html", body)
	m.Embed(imgPath)
	//m.SetAddressHeader("Cc", "skerlight@naver.com", "Song")
	d := gomail.NewDialer(host, port, sender, pw)

	if err := d.DialAndSend(m); err != nil {
		reqLogger.V(1).Error(err, " Sent Mail to User [ "+recipient+"] Failed")
		panic(err)
	}
	reqLogger.V(3).Info(" Sent Mail to User [ " + recipient + " ]")
}

func RemoveValue(slice []string, value string) []string {
	temp := []string{}
	for i := 0; i < len(slice); i++ {
		if slice[i] != value {
			temp = append(temp, slice[i])
		}
	}
	return temp
}

func UpdateResourceList(reqLogger logr.Logger) {
	ResourceList = []string{}
	apiGroupList := &metav1.APIGroupList{}
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/").DoRaw(context.TODO())
	if err != nil {
		reqLogger.V(1).Error(err, "Failed to get api group from k8s")
	}
	if err := json.Unmarshal(data, apiGroupList); err != nil {
		reqLogger.V(1).Error(err, "Failed to unmarshal api group")
	}

	for _, apiGroup := range apiGroupList.Groups {
		for _, version := range apiGroup.Versions {
			apiResourceList := &metav1.APIResourceList{}
			path := strings.Replace("/apis/{GROUPVERSION}", "{GROUPVERSION}", version.GroupVersion, -1)
			data, err := Clientset.RESTClient().Get().AbsPath(path).DoRaw(context.TODO())
			if err != nil {
				reqLogger.V(1).Error(err, "Failed to get api reousrce list from k8s")
			}
			if err := json.Unmarshal(data, apiResourceList); err != nil {
				reqLogger.V(1).Error(err, "Failed to unmarshal api resource list")
			}

			for _, apiResource := range apiResourceList.APIResources {
				fullName := version.GroupVersion
				idx := strings.Index(fullName, "/")
				fullName = fullName[:idx]
				fullName += "/"
				fullName += apiResource.Name
				ResourceList = append(ResourceList, fullName)
			}
		}
	}
}
