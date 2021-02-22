node {
    def gitHubBaseAddress = "github.com"
    def goHome = "/usr/local/go/bin"
    def buildDir = "/var/lib/jenkins/workspace/hypercloud-go-operator"
    
    def scriptHome = "${buildDir}/scripts"
	
	def gitAddress = "${gitHubBaseAddress}/tmax-cloud/hypercloud-single-operator.git"

	def version = "${params.majorVersion}.${params.minorVersion}.${params.tinyVersion}.${params.hotfixVersion}"
	def preVersion = "${params.preVersion}"
	
	def imageTag = "b${version}"
				
	def userName = "dnxorjs1"
	def userEmail = "taegeon_woo@tmax.co.kr"
    
    stage('git pull') {
        dir(buildDir){
            git branch: "${params.buildBranch}",
            credentialsId: '${userName}',
            url: "http://${gitAddress}"

            // git pull
            sh "git checkout ${params.buildBranch}"
            sh "git config --global user.name ${userName}"
            sh "git config --global user.email ${userEmail}"
            sh "git config --global credential.helper store"
        
            sh "git fetch --all"
            sh "git reset --hard origin/${params.buildBranch}"
            sh "git pull origin ${params.buildBranch}"

            sh '''#!/bin/bash
            export PATH=$PATH:/usr/local/go/bin
            export GO111MODULE=on
            go build -o bin/manager main.go
            '''
        }
    }
    
    stage('make manifests') {
	    sh "sed -i 's#{imageTag}#${imageTag}#' ./config/manager/kustomization.yaml"
        sh "sudo kubectl kustomize ./config/default/ > bin/hypercloud-go-operator-v${version}.yaml"
        sh "sudo kubectl kustomize ./config/crd/ > bin/crd-v${version}.yaml"
        sh "sudo tar -zvcf bin/hypercloud-go-operator-manifests-v${version}.tar.gz bin/hypercloud-go-operator-v${version}.yaml bin/crd-v${version}.yaml"
        
        sh "sudo mkdir -p build/manifests/v${version}"
        sh "sudo cp bin/*v${version}.yaml build/manifests/v${version}/"
    }

    stage('image build/push') {
        sh "sudo docker build --tag tmaxcloudck/hypercloud-single-operator:${imageTag} ."
        sh "sudo docker push tmaxcloudck/hypercloud-single-operator:${imageTag}"
        sh "sudo docker rmi tmaxcloudck/hypercloud-single-operator:${imageTag}"
    }

    stage('make-changelog') {
        sh "echo targetVersion: ${version}, preVersion: ${preVersion}"
        sh "sudo sh ${scriptHome}/make-changelog.sh ${version} ${preVersion}"
    }

    stage('gitcommit & push') {
        dir("${buildDir}") {
			sh "git checkout ${params.buildBranch}"
            sh "git add -A"
			sh "git reset ./config/manager/kustomization.yaml"
            def commitMsg = "[Distribution] Release commit for hypercloud-single-operator v${version}"
            sh (script: "git commit -m \"${commitMsg}\" || true")
            sh "git tag v${version}"
			sh "sudo git push -u origin +${params.buildBranch}"
            sh "sudo git push origin v${version}"
            
            sh "git fetch --all"
			sh "git reset --hard origin/${params.buildBranch}"
			sh "git pull origin ${params.buildBranch}"
        }
    }
    
    // stage('release') {
    //     withCredentials([usernamePassword(credentialsId: 'hypercloud-bot', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
    //         def body = '\\{\\"tag_name\\":\\"' + "v${version}"+ '\\",\\"name\\":\\"' + "v${version}" + '\\",\\"body\\":\\"test\\"\\}'
    //         def releaseResult = sh returnStdout: true, script: "curl -u ${USERNAME}:${PASSWORD} -H \"Content-Type: application/vnd.github.v3+json\" -d ${body} -X POST https://api.github.com/repos/tmax-cloud/hypercloud-single-operator/releases | jq '.id' | tr -d '\n' "
            
    //         def filename = "hypercloud-manifests-v${version}.tar.gz"
    //         sh "curl -u ${USERNAME}:${PASSWORD} -H \"Content-Type: application/zip\" --data-binary @bin/${filename} -X POST https://uploads.github.com/repos/tmax-cloud/hypercloud-single-operator/releases/${releaseResult}/assets?name=${filename}"
    //     }
    // }
    
//     stage('clean repo') {
//         sh "sudo rm -rf ${buildDir}/*"
//     }
    
//     stage('send email') {
//         def dateFormat = new SimpleDateFormat("yyyy.MM.dd E")
//         def date = new Date()
                
//         def today = dateFormat.format(date) 
        
//         emailext (
//             subject: "Release hypercloud-single-operator v${version}",
//             body: 
// """
// 안녕하세요. ck2-3팀입니다.
// hypercloud-single-operator 정기 배포 안내 메일입니다.

// 배포 관련 아래 링크를 확인 부탁드립니다.
// https://github.com/tmax-cloud/hypercloud-single-operator/releases/tag/v${version}

// 감사합니다.

// ===

// ${today}
// Hypercloud-go-operator 배포
//     * HyperCloudServer
//         * version: v${version}
//         * image: docker.io/tmaxcloudck/hypercloud-single-operator:v${version}
        
// """,
//             to: "jaihwan_jung@tmax.co.kr;jaehyan1013@naver.com",
//             from: "hypercloudbot@gmail.com"
//         )
//     }
}
