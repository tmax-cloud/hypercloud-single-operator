package util

const (
	RBAC_API_GROUP = "rbac.authorization.k8s.io"
	// HYPERCLOUD_NAMESPACE                = "hypercloud4-system"
	HYPERCLOUD_NAMESPACE                = "hypercloud-system"
	DEFAULT_NETWORK_POLICY_CONFIG_MAP   = "default-networkpolicy-configmap"
	NETWORK_POLICY_YAML                 = "networkpolicies.yaml"
	TRIAL_SUCCESS_CONFIRM_MAIL_CONTENTS = "<!DOCTYPE html>\r\n" +
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
	TRIAL_FAIL_CONFIRM_MAIL_CONTENTS = "<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"    <meta charset=\"UTF-8\">\r\n" +
		"    <title>HyperCloud 서비스 신청 결과 알림</title>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<div style=\"border: #c5c5c8 0.06rem solid; border-bottom: 0; width: 42.5rem; height: 43.19rem; padding: 0 1.25rem\">\r\n" +
		"    <header>\r\n" +
		"        <div style=\"margin: 0;\">\r\n" +
		"            <p style=\"font-size: 1rem; font-weight: bold; color: #333333; line-height: 3rem; letter-spacing: 0; border-bottom: #c5c5c8 0.06rem solid;\">\r\n" +
		"                HyperCloud 서비스 신청 결과 알림\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </header>\r\n" +
		"    <section>\r\n" +
		"        <figure style=\"text-align: center;\">\r\n" +
		"            <img style=\"margin: 0.94rem 0;\"\r\n" +
		"                 src=\"cid:trial-disapproval\">\r\n" +
		"        </figure>\r\n" +
		"        <div style=\"width: 35.70rem; margin: 2rem 2.27rem;\">\r\n" +
		"            <p style=\"line-height: 1.38rem;\">\r\n" +
		"                안녕하세요? TmaxCloud 입니다. <br>\r\n" +
		"                TmaxCloud에 관심을 가져 주셔서 감사합니다. 고객님의 서비스 신청을 검토하였으며, <br>\r\n" +
		"                그 결과로 고객님의 서비스 신청이 승인되지 않았음을 알려드립니다. <br>\r\n" +
		"                비승인 사유는 아래와 같습니다. <br>\r\n" +
		"                <br>\r\n" +
		"                <span style=\"font-weight: 600;\">비승인 사유</span>\r\n" +
		"                <p>%%FAIL_REASON%%</p>\r\n" +
		"                <br>\r\n" +
		"                <br>\r\n" +
		"                감사합니다. <br>\r\n" +
		"                TmaxCloud 드림.\r\n" +
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
	VERIFY_MAIL_CONTENTS = "<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"    <meta charset=\"UTF-8\">\r\n" +
		"    <title>이메일을 인증해주세요.</title>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<div style=\"border: #c5c5c8 0.06rem solid; border-bottom: 0; width: 42.5rem; height: 50.94rem; padding: 0 1.25rem\">\r\n" +
		"    <header>\r\n" +
		"        <div style=\"margin: 0;\">\r\n" +
		"            <p style=\"font-size: 1rem; font-weight: bold; color: #333333; line-height: 3rem; letter-spacing: 0; border-bottom: #c5c5c8 0.06rem solid;\">\r\n" +
		"                [인증번호 : @@verifyNumber@@] 이메일을 인증해 주세요.\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </header>\r\n" +
		"    <section>\r\n" +
		"        <figure style=\"text-align: center;\">\r\n" +
		"            <img style=\"margin: 2.38rem 0;\"\r\n" +
		"                 src=\"cid:index\">\r\n" +
		"        </figure>\r\n" +
		"        <div style=\"width: 27.06rem; margin: 0 7.70rem;\">\r\n" +
		"            <p style=\"font-size: 1.25rem; font-weight: bold; line-height: 3rem;\">\r\n" +
		"                인증번호 @@verifyNumber@@\r\n" +
		"            </p>\r\n" +
		"            <p style=\"line-height: 1.38rem;\">\r\n" +
		"                안녕하세요? <br>\r\n" +
		"                TmaxCloud를 이용해 주셔서 감사합니다. <br>\r\n" +
		"                가입 화면에서 인증번호를 입력해 주세요. <br>\r\n" +
		"                감사합니다. <br>\r\n" +
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
	TRIAL_TIME_OUT_CONTENTS = "<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"    <meta charset=\"UTF-8\">\r\n" +
		"    <title>HyperCloud 서비스 기간 만료 안내 알림</title>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<div style=\"border: #c5c5c8 0.06rem solid; border-bottom: 0; width: 42.5rem; height: 43.19rem; padding: 0 1.25rem\">\r\n" +
		"    <header>\r\n" +
		"        <div style=\"margin: 0;\">\r\n" +
		"            <p style=\"font-size: 1rem; font-weight: bold; color: #333333; line-height: 3rem; letter-spacing: 0; border-bottom: #c5c5c8 0.06rem solid;\">\r\n" +
		"                HyperCloud 서비스 기간 만료 안내 알림\r\n" +
		"            </p>\r\n" +
		"        </div>\r\n" +
		"    </header>\r\n" +
		"    <section>\r\n" +
		"        <figure style=\"text-align: center;\">\r\n" +
		"            <img style=\"margin: 2.38rem 0;\"\r\n" +
		"                 src=\"cid:service-timeout\">\r\n" +
		"        </figure>\r\n" +
		"        <div style=\"width: 34.44rem; margin: 0 4rem;\">\r\n" +
		"<!--            <p style=\"font-size: 1.25rem; font-weight: bold; line-height: 3rem;\">-->\r\n" +
		"<!--                인증번호 1256-->\r\n" +
		"<!--            </p>-->\r\n" +
		"            <p style=\"line-height: 1.38rem;\">\r\n" +
		"                안녕하세요? <br>\r\n" +
		"                TmaxCloud를 이용해 주셔서 감사합니다. <br>\r\n" +
		"                고객님께서 사용중인 Trial 서비스가 <span style=\"color: #F26868;\">%%TRIAL_END_TIME%%</span>에 만료됩니다. <br>\r\n" +
		"                Trial 서비스 이용 만료 시 사용중인 네임스페이스의 리소스는 모두 삭제됩니다. <br>\r\n" +
		"                Trial 서비스 만료 기간 이전에 <span style=\"color: #187EE3;\">유료 서비스로 전환</span> 혹은 리소스를 백업해주시기 바랍니다. <br>\r\n" +
		"                <br>\r\n" +
		"                감사합니다. <br>\r\n" +
		"                TmaxCloud 드림.\r\n" +
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
