#!/usr/bin/python3.10
# -*- coding: utf-8 -*-
import json
import ddddocr
import requests
def gettoken():
# cnncvd官网登陆认证地址
    urlLogin = "https://www.cnnvd.org.cn/web/login"
# cnnvd官网获取登录验证码图片地址
    urlImage = "https://www.cnnvd.org.cn/web/verificationCode/getBase64Image"
# 获取验证码图片（base64）
    response = requests.get(url=urlImage,verify=False)
# json序列化
    imageData = json.loads(response.text)
# 提取图片数据
    image =imageData["data"]["image"][22:]
#print(image)
    ocr = ddddocr.DdddOcr()
# 图片识别
    code = ocr.classification(image)
    #print(code)
# 提取辅助验证信息
    verifyToken =imageData["data"]["verifyToken"]
    #print(verifyToken)
# 登录信息
    postData ={
        "username": "kitsdk@163.com",
        "password": "bed128365216c019988915ed3add75fb",
        "code": code,
        "verifyToken": verifyToken
        }
# 模拟登录
    rq=requests.post(url=urlLogin,data=json.dumps(postData),headers={'Content-Type':'application/json'},verify=False)
# 序列化返回信息
    loginData = json.loads(rq.text)
#提取token
    print(loginData["data"]["token"])
    #return (loginData["data"]["token"])

if __name__ == '__main__' :
    gettoken()
