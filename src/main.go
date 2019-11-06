package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"
)

type WechatAuth struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type EggOrder struct {
	OrderId string
	Client  string
	Goods   string
	Address string
	Amount  string
}

func main() {
	// 从本地读取环境变量
	godotenv.Load()

	host := os.Getenv("host")
	loginUrl := host + "/login/login.jsp"
	loginData := map[string]string{"loginid": os.Getenv("username"), "password": os.Getenv("password")}
	//实例化爬虫对象
	c := colly.NewCollector()

	//登录
	eggLogin(c, loginUrl, loginData)

	//开始抓取订单列表页面，并且确认所有订单
	eggList(c, getToConfirmOrderUrl())

	//获取微信token，用于发送模板消息
	//token := getAccessToken()

	//对订单进行依次发送模板消息
}

func eggLogin(c *colly.Collector, loginUrl string, loginData map[string]string) *colly.Collector {
	err := c.Post(loginUrl, loginData)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func eggList(c *colly.Collector, listUrl string) {
	// 绑定回调事件，需要注意登录此时已经完成，不会触发任何回调
	c.OnResponse(func(r *colly.Response) {
		log.Println("页面请求成功：", String(r.Body)[0:10])
		//log.Println("response received", r.StatusCode)
	})
	// 抓取具体的HTML元素回调
	c.OnHTML("input[type=checkbox]", func(e *colly.HTMLElement) {
		order_id, _ := strconv.Atoi(e.Attr("value"))
		eggConfirmOrder(c, order_id, e) //确认订单
	})
	err := c.Visit(listUrl)
	if err != nil {
		log.Fatal("抓取待确认订单失败！")
	}
}

func eggConfirmOrder(c *colly.Collector, order_id int, e *colly.HTMLElement) {
	//只有id数字五位数，才有可能是正常订单，其余都是噪音
	if order_id > 100 {
		fmt.Print("订单号："+e.Attr("value"), "\n")
		//发送POST确认订单
		//err = c.Post(host+"/renovation/web/ffzpub/direct-order/deliver.jsp", map[string]string{"id": e.Attr("value")})
	}
}

//拼接得到订单列表页面URL
func getToConfirmOrderUrl() string {
	host := os.Getenv("host")
	toConfirmOrderUrl := host + "/renovation/web/ffzpub/direct-order/search.jsp?"
	query := "&starttimes=&endtimes=&customname=&ordernum=&orderstatus=&ordertype=&salename=&opname=&fixarea=&maintain=&dstoreid=&queryTime="
	t := time.Now().AddDate(0, 0, +1)
	tomorrow := t.Format("2006-01-02")
	toConfirmOrderUrl += "startdelivar=" + tomorrow
	toConfirmOrderUrl += "&stopdelivar=" + tomorrow
	toConfirmOrderUrl += query
	return toConfirmOrderUrl
}

func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func getAccessToken() string {
	wechatAccessTokenUrl := os.Getenv("wechatAccessTokenUrl")
	auth := WechatAuth{}
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		err := json.Unmarshal(r.Body, &auth);
		if err != nil {
			log.Print("获取AccessToken失败")
		}
	})
	err := c.Visit(wechatAccessTokenUrl)
	if err != nil {
		log.Print("访问AccessToken失败")
	}
	return auth.AccessToken
}
