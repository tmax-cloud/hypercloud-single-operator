package util

import (
	"context"
	"strconv"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/gomail.v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetTrialNSTimer(ns *v1.Namespace, client client.Client, reqLogger logr.Logger) {
	reqLogger.Info("[Trial Timer] TrialNSTimer for Trial NS[ " + ns.Name + " ] Set Service Start")
	currentTime := time.Now()
	createTime := ns.CreationTimestamp.Time
	reqLogger.Info("[Trial Timer] CreateTime of Trial NS[ " + ns.Name + " ] : " + createTime.String())
	mailTime := createTime.AddDate(0, 0, 23)
	// mailTime := createTime.Add(time.Second * 23)  // for test
	deleteTime := createTime.AddDate(0, 0, 30)
	// deleteTime := createTime.Add(time.Second * 30)  // for test

	if ns.Labels["period"] != "" {
		period, _ := strconv.Atoi(ns.Labels["period"])
		deleteTime = createTime.AddDate(0, 0, period*30)
		mailTime = deleteTime.AddDate(0, 0, -7)
	}
	if mailTime.After(currentTime) {
		time.AfterFunc(time.Duration((mailTime.UnixNano() - currentTime.UnixNano())), func() {
			reqLogger.Info(" [Trial Timer] Trial NameSpace [ " + ns.Name + " ] Mail Service before 1 weeks of deletion Start")
			nsFound := &v1.Namespace{}
			if err := client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, nsFound); err != nil && errors.IsNotFound(err) {
				reqLogger.Info(" [Trial Timer]  NameSpace [ " + ns.Name + " ] has Deleted, Nothing to do")
			}
			if nsFound.Labels != nil && nsFound.Labels["trial"] != "" && nsFound.Annotations != nil && nsFound.Annotations["owner"] != "" {
				reqLogger.Info(" [Trial Timer] Still Trial NameSpace, Send Info Mail to User [ " + nsFound.Annotations["owner"] + " ]")
				subject := " 신청해주신 Trial NameSpace [ " + nsFound.Name + " ] 만료 안내 "
				body := TRIAL_TIME_OUT_CONTENTS
				body = strings.ReplaceAll(body, "%%TRIAL_END_TIME%%", deleteTime.Format("2006-01-02"))
				SendMail(nsFound.Annotations["owner"], subject, body, "/home/tmax/hypercloud4-operator/_html/img/service-timeout.png", "service-timeout", reqLogger)
			} else {
				reqLogger.Info(" [Trial Timer] Paid NameSpace, Nothing to do")
			}

		})
		reqLogger.Info(" [Trial Timer] Set Trial NameSpace Sending Mail Timer Success ")
		reqLogger.Info(" [Trial Timer] MailSendTime for Trial NS[ " + ns.Name + " ] : " + mailTime.String())

		ns.Labels["mailSendDate"] = mailTime.Format("2006-01-02")
	} else {
		reqLogger.Info(" [Trial Timer] Mail for Alert Deletion for This Trial Namespace [" + ns.Name + "] already Sent to " + ns.Annotations["owner"])
	}

	if deleteTime.After(currentTime) {
		time.AfterFunc(time.Duration((deleteTime.UnixNano() - currentTime.UnixNano())), func() {
			reqLogger.Info(" [Trial Timer] Trial NameSpace [ " + ns.Name + " ] deletion Start")
			nsFound := &v1.Namespace{}
			if err := client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, nsFound); err != nil && errors.IsNotFound(err) {
				reqLogger.Error(err, " [Trial Timer]  NameSpace [ "+ns.Name+" ] has Deleted, Nothing to do")
			}
			if nsFound.Labels != nil && nsFound.Labels["trial"] != "" && nsFound.Annotations != nil && nsFound.Annotations["owner"] != "" {
				reqLogger.Info(" [Trial Timer] Still Trial NameSpace, Delete Expired Namespace [ " + nsFound.Name + " ]")
				if err := client.Delete(context.TODO(), nsFound); err != nil {
					reqLogger.Error(err, " [Trial Timer] Failed to Delete NameSpace [ "+ns.Name+" ]")
					panic(err)
				} else if err := client.Delete(context.TODO(), &rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "CRB-" + nsFound.Name,
					},
				}); err != nil {
					reqLogger.Error(err, " [Trial Timer] Failed to Delete ClusterRoleBinding [ "+"CRB-"+nsFound.Name+" ]")
					panic(err)
				} else {
					reqLogger.Info(" [Trial Timer] Delete Expired Namespace [ " + nsFound.Name + " ] Success")
				}
			} else {
				reqLogger.Info(" [Trial Timer] Paid NameSpace, Nothing to do")
			}
		})
		reqLogger.Info(" [Trial Timer] Set Trial NameSpace delete Timer Success ")
		reqLogger.Info(" [Trial Timer] DeletionTime for Trial NS[ " + ns.Name + " ] : " + deleteTime.String())

		ns.Labels["deletionDate"] = deleteTime.Format("2006-01-02")
		if err := client.Update(context.TODO(), ns); err != nil {
			reqLogger.Error(err, "[Trial Timer] Replace NameSpace for new Label Failed")
			panic(err)
		} else {
			reqLogger.Info(" [Trial Timer] Replace NameSpace for new Label Success ")
		}
	} else {
		reqLogger.Info(" [Trial Timer] This Trial Namespace [" + ns.Name + "] has Already Expired, Check Why This NameSpace is Still Exists")
	}
}

func SendMail(recipient string, subject string, body string, imgPath string, imgCid string, reqLogger logr.Logger) {
	reqLogger.Info(" Send Mail to User [ " + recipient + "] Start")
	host := "mail.tmax.co.kr"
	port := 25
	sender := "no-reply-tc@tmax.co.kr"
	pw := "!@tcdnsdudxla11"

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	//m.Attach("/home/Alex/lolcat.jpg")
	//m.SetAddressHeader("Cc", "skerlight@naver.com", "Song")
	d := gomail.NewDialer(host, port, sender, pw)
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
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
