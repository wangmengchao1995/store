package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"bj2qFresh/models"
	"github.com/astaxie/beego/utils"
	"strconv"
	"github.com/gomodule/redigo/redis"
)

type UserController struct {
	beego.Controller
}

func(this*UserController)ShowRegister(){
	this.TplName = "register.html"
}

//处理注册数据
func(this*UserController)HandleRegister(){
	//获取数据
	userName := this.GetString("user_name")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")
	//校验数据
	if userName == "" || pwd == "" || cpwd == "" || email == ""{
		this.Data["errmsg"] = "数据不能为空，请重新注册!"
		this.TplName = "register.html"
		return
	}
	//邮箱格式校验
	reg,_ :=regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res := reg.FindString(email)
	if res == "" {
		this.Data["errmsg"] = "邮箱格式不正确，请重新注册!"
		this.TplName = "register.html"
		return
	}
	//两次输入密码是否一直
	if pwd != cpwd {
		this.Data["errmsg"] = "两次密码输入不正确，请重新注册！"
		this.TplName = "register.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	//获取插入对象
	var user models.User
	//给插入对象赋值
	user.Name = userName
	user.PassWord = pwd
	user.Email = email
	_,err :=o.Insert(&user)
	if err != nil{
		this.Data["errmsg"] = "注册失败，请重新注册！"
		this.TplName = "register.html"
		return
	}

	//邮箱激活
	emailConfig := `{"username":"563364657@qq.com","password":"vevproxudtibbbja","host":"smtp.qq.com","port":587}`
	ems := utils.NewEMail(emailConfig)
	ems.From = "563364657@qq.com"
	ems.To = []string{email}
	ems.Subject = "天天生鲜用户激活"
	//ems.HTML = "<a href=\"http://192.168.110.81:8080/active?id="+strconv.Itoa(user.Id)+">点击该链接，天天生鲜用户激活</a>"
	ems.HTML = "<a href=\"http://192.168.110.81:8080/active?id="+strconv.Itoa(user.Id)+"\">点击该链接，天天生鲜用户激活</a>"

	err = ems.Send()
	beego.Error(err)
	//返回数据
	this.Ctx.WriteString("注册成功,请去目标邮箱激活！")
}

//处理激活业务
func(this*UserController)HandleActive(){
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err!=nil{
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return

	}
	//操作数据
	//更新操作  id
	//获取orm对象
	o := orm.NewOrm()
	//获取更新对象
	var user models.User
	//给更新对象复制
	user.Id = id
	//查询
	err = o.Read(&user)
	if err != nil{
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return
	}
	//给要更新的字段赋新值
	user.Active = true
	_,err = o.Update(&user)
	if err != nil{
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return
	}


	//返回数据
	this.Redirect("/login",302)
}

//展示登录界面
func(this*UserController)ShowLogin(){
	userName := this.Ctx.GetCookie("userName")
	if userName == ""{
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}else {
		this.Data["userName"] = userName
		this.Data["checked"] = "checked"
	}

	this.TplName = "login.html"
}
//处理登录业务
func(this*UserController)HandleLogin(){
	//获取数据
	userName := this.GetString("username")
	pwd := this.GetString("pwd")
	//校验数据
	if userName == "" || pwd ==""{
		this.Data["errmsg"] = "用户名或者密码不能为空"
		this.TplName = "login.html"
		return
	}
	//处理数据
	//查询
	//获取orm对象
	o := orm.NewOrm()
	//获取查询兑现g
	var user models.User
	//给查询条件赋值
	user.Name = userName
	//查询
	err := o.Read(&user,"Name")
	if err != nil{
		this.Data["errmsg"] = "用户不存在，请重新登录"
		this.TplName = "login.html"
		return
	}
	//密码校验
	if user.PassWord != pwd{
		this.Data["errmsg"] = "用户密码错误，请重新登录"
		this.TplName = "login.html"
		return
	}
	//用户是否激活
	if user.Active == false{
		this.Data["errmsg"] = "用户未激活，请先激活用户"
		this.TplName = "login.html"
		return
	}

	//在勾选的情况下登陆成功，记住用户名
	check := this.GetString("check")
	if check == "on"{
		this.Ctx.SetCookie("userName",userName,3600)
	}else{
		this.Ctx.SetCookie("userName",userName,-1)
	}

	//设置session
	this.SetSession("userName",userName)

	//返回数据
	this.Redirect("/",302)
}

//退出登录
func(this*UserController)Logout(){
	//删除session
	this.DelSession("userName")
	//跳转页面
	this.Redirect("/login",302)
}

func ShowLayout(this*UserController){
	//获取userName
	userName := this.GetSession("userName")

	this.Data["userName"] = userName.(string)

	this.Layout = "usercenterLayout.html"
}

//展示用户中心信息页
func(this*UserController)ShowUserCenterInfo(){
	//从session中获取用户名
	ShowLayout(this)

	//获取当前用户的默认地址
	o := orm.NewOrm()
	//指定要查询的表
	qs := o.QueryTable("Address")
	//关联用户表
	qs = qs.RelatedSel("User")
	//判断当前用户
	userName := this.GetSession("userName")
	qs = qs.Filter("User__Name",userName.(string))
	//获取默认地址
	qs = qs.Filter("Isdefault",true)
	//把查询到的数据，放到容器里面
	var address models.Address
	err := qs.One(&address)
	if err != nil{
		this.Data["address"] = ""
	}else {
		this.Data["address"] = address
	}
	//获取历史浏览记录
	conn,err := redis.Dial("tcp","192.168.110.81:6379")
	if err != nil{
		beego.Error("redis链接失败",err)
	}
	defer conn.Close()
	//userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user,"Name")
	resp,err := conn.Do("lrange","history_"+strconv.Itoa(user.Id),0,4)
	//回复助手函数
	goodsId,err := redis.Ints(resp,err)
	if err != nil{
		beego.Error("redis获取商品错误",err)
	}
	//beego.Info(goodsId)
	var goodsSku []models.GoodsSKU
	for _,id := range goodsId{
		var goods models.GoodsSKU
		goods.Id = id
		o.Read(&goods)
		goodsSku = append(goodsSku, goods)
	}
	this.Data["goodsSkus"] = goodsSku





	this.TplName = "user_center_info.html"
}

//展示用户中心订单页
func(this*UserController)ShowUserCenterOrder(){
	//调用试图布局
	ShowLayout(this)
	this.TplName = "user_center_order.html"
}

//用户中心地址页
func(this*UserController)ShowUserCenterSite(){
	ShowLayout(this)

	//查询当前用户的默认地址
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var address models.Address
	//查询
	userName := this.GetSession("userName")
	//select * from address where user.name = userName and Isdefault = true
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName.(string)).Filter("Isdefault",true).One(&address)

	this.Data["address"] = address

	this.TplName = "user_center_site.html"
}

//处理用户中心地址页数据
func(this*UserController)HandleUserCenterSite(){
	//获取数据
	recever := this.GetString("recever")
	addr := this.GetString("addr")
	zipCode := this.GetString("zipCode")
	phone := this.GetString("phone")
	//校验数据
	if recever == "" || addr == "" || zipCode == "" || phone == ""{
		beego.Error("添加地址页面，获取数据失败")
		this.Redirect("/user/userCenterSite",302)
		return
	}

	//处理数据
	//插入操作
	//获取orm对象
	o := orm.NewOrm()
	//获取插入对象
	var address models.Address
	//给插入对象赋值
	address.Receiver = recever
	address.Zipcode = zipCode
	address.Addr = addr
	address.Phone = phone


	//一对多的插入
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user,"Name")

	address.User = &user
	//判断当前用户是否有默认地址，如果没有,则直接插入默认地址，
	// 如果有默认地址，把默认地址更新为非默认地址，把新插入的地址设置为默认地址
	//获取当前用户的默认地址
	var oldAddress models.Address
	err := o.QueryTable("Address").RelatedSel("User").Filter("User__Id",user.Id).Filter("Isdefault",true).One(&oldAddress)
	if err != nil{
		address.Isdefault = true
	}else {
		oldAddress.Isdefault = false
		//更新操作
		o.Update(&oldAddress)
		address.Isdefault = true
	}


	//插入
	o.Insert(&address)

	//返回数据
	this.Redirect("/user/userCenterSite",302)
}
