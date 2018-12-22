package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"bj2qFresh/models"
	"math"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type GoodsController struct {
	beego.Controller
}
//展示首頁內容
func(this*GoodsController)ShowIndex(){
	//当登录的时候显示欢迎你，username,当没有登录显示登录注册
	userName := this.GetSession("userName")
	if userName == nil{
		this.Data["userName"] = ""
	}else {
		this.Data["userName"] = userName.(string)
	}

	//獲取相應數據
	//獲取orm對象
	o := orm.NewOrm()
	//獲取查詢對象
	var goodstypes []models.GoodsType
	//查詢
	o.QueryTable("GoodsType").All(&goodstypes)

	//查詢獨享
	var indexGoodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanners)

	//查詢對象  活動推廣
	var indexPromotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexPromotionBanners)
	//查詢首頁展示商品
	var goodsSkus = make([]map[string]interface{},len(goodstypes))

	//把類型對象放入我們的map容器中
	for index,_ := range goodsSkus{
		temp := make(map[string]interface{})
		temp["types"] = goodstypes[index]
		goodsSkus[index] = temp
	}
	//存商品數據
	for _,goodsMap := range goodsSkus{
		var goodsImage []models.IndexTypeGoodsBanner
		var goodsText []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSku").Filter("GoodsType",goodsMap["types"]).Filter("DisplayType",0).All(&goodsText)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSku").Filter("GoodsType",goodsMap["types"]).Filter("DisplayType",1).All(&goodsImage)

		goodsMap["goodsImage"] = goodsImage
		goodsMap["goodsText"] = goodsText
	}

	this.Data["goodsSkus"] = goodsSkus


	this.Data["goodsTypes"] = goodstypes
	this.Data["indexGoodsBanners"] = indexGoodsBanners
	this.Data["indexPromotionBanners"] = indexPromotionBanners
	beego.Info(goodsSkus)

	this.TplName = "index.html"
}

//展示Layout页面
func ShowGoodsLayout(this*GoodsController, typeId int){
	//获取类型数据
	//查询类型
	o := orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes
	//获取新品数据
	//获取同一类型的新品数据
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"] = newGoods
	this.Layout = "goodsLayout.html"
}

//展示商品详情页
func(this*GoodsController)ShowGoodsDetail(){
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err != nil{
		beego.Error("请求路径错误")
	}
	//处理数据
	//查询
	o := orm.NewOrm()
	//获取查询对象
	var goodsSku models.GoodsSKU
	//给查询条件赋值
	goodsSku.Id = id
	//查询
	err = o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goodsSku)
	if err != nil{
		beego.Error("查询商品数据错误")
	}
	//goodsSku.GoodsType.Id

	ShowGoodsLayout(this,goodsSku.GoodsType.Id)


	//添加历史浏览记录
	//判断是否是登录状体
	userName := this.GetSession("userName")
	if userName != nil{
		//需要获取存储的信息  用户id 和商品id
		//1.获取当前用户信息
		var user models.User
		user.Name = userName.(string)
		o.Read(&user,"Name")

		//存储
		//链接，获取redis操作对象
		conn,err := redis.Dial("tcp","192.168.110.81:6379")
		defer conn.Close()
		if err != nil{
			beego.Error("redis链接失败")
		}
		conn.Do("lrem","history_"+strconv.Itoa(user.Id),0,id)
		conn.Do("lpush","history_"+strconv.Itoa(user.Id),id)

	}


	//返回数据
	this.Data["goodsSku"] = goodsSku
	this.TplName = "detail.html"
}

//实现分页
func PageEdior(pageCount float64,pageIndex int)([]int){
	//判断显示哪些页码
	var pages []int

	if pageCount <= 5{
		pages = make([]int,int(pageCount))
		i := 1
		for pageCount > 0{
			pages[i-1] = i
			pageCount -= 1
			i += 1
		}
	}else  if pageIndex <= 3{
		pages = make([]int,5)
		i := 1
		//当前页码等于
		var temp = 5
		for temp > 0{
			pages[i-1] = i
			temp -= 1
			i += 1
		}
	}else if pageIndex >= int(pageCount) - 2{
		pages = make([]int,5)
		//给后三页赋值
		temp := 5
		i := 1
		for temp >0{
			pages[i-1] = int(pageCount) - temp + 1
			temp -= 1
			i +=1
		}
	}else {
		pages = make([]int,5)
		temp := 2
		i := 1
		for temp > -3{
			pages[i-1] = pageIndex - temp
			temp -= 1
			i += 1
		}
		beego.Info(pages)
	}
	return pages
}


//展示商品列表页
func(this*GoodsController)ShowGoodsList(){
	//获取数据
	typeId ,err := this.GetInt("id")
	//校验数据
	if err != nil{
		beego.Error("获取商品类型错误")
	}
	//处理数据
	//获取和传递过来的类型一直的商品
	o := orm.NewOrm()
	//获取查询对象
	var goodsSkus []models.GoodsSKU
	//查询  默认排序
	//o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).All(&goodsSkus)
	//this.Data["goodsSkus"] = goodsSkus
	//按照相应的排序方式获取数据
	sort := this.GetString("sort")



	this.Data["sort"] = sort
	ShowGoodsLayout(this,typeId)

	//实现分页
	//分析讨论，现在应该展示哪些页码
	//获取总页码
	count,_ :=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).Count()
	pageSize := 1
	pageCount := math.Ceil(float64(count)/float64(pageSize))

	pageIndex,err := this.GetInt("pageIndex")
	if err != nil{
		pageIndex = 1
	}

	//判断显示哪些页码
	var pages []int
	pages = PageEdior(pageCount,pageIndex)

	this.Data["pages"] = pages
	start := (pageIndex - 1) * pageSize
	if sort == "price"{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("Price").Limit(pageSize,start).All(&goodsSkus)
	}else if sort == "sale"{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("Sales").Limit(pageSize,start).All(&goodsSkus)
	}else{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).Limit(pageSize,start).All(&goodsSkus)
	}
	this.Data["goodsSkus"] = goodsSkus

	prePage := pageIndex - 1
	if prePage < 1{
		prePage = 1
	}
	nextPage := pageIndex + 1
	if nextPage > int(pageCount){
		nextPage = int(pageCount)
	}
	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	this.Data["pageIndex"] = pageIndex





	//返回数据
	this.Data["typeId"] = typeId
	this.TplName = "list.html"
}

//商品搜索
func(this*GoodsController)HandleSearch(){
	//获取数据
	search := this.GetString("searchName")
	o := orm.NewOrm()
	var goodsSkus []models.GoodsSKU
	//校验数据
	if search == ""{
		o.QueryTable("GoodsSKU").All(&goodsSkus)
	}else {
		o.QueryTable("GoodsSKU").Filter("Name__icontains",search).All(&goodsSkus)
	}

	//处理数据

	//返回数据
	this.Data["search"]=goodsSkus
	this.Layout = "goodsLayout.html"
	this.TplName = "search.html"
}

