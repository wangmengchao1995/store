package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"bj2qFresh/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

func(this*OrderController)ShowOrder(){
	//获取数据
	//id,_ := this.GetInt("id")
	ids :=this.GetStrings("id")
	beego.Info("获取到的id",ids)
	//校验数据
	if len(ids) == 0{
		beego.Error("请求路径错误")
	}


	//处理数据
	//获取当前用户的所有地址信息
	o := orm.NewOrm()
	var adds []models.Address
	//获取用户
	userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user,"Name")

	o.QueryTable("Address").RelatedSel("User").Filter("User",user).All(&adds)

	this.Data["adds"] = adds
	//商品数据
	//商品数量
	conn,_ :=redis.Dial("tcp","192.168.110.81:6379")

	goods := make([]map[string]interface{},0)
	totalPrice := 0
	totalCount := 0
	for _,value:= range ids{
		temp := make(map[string]interface{})
		var goodsSku models.GoodsSKU
		id ,_:=strconv.Atoi(value)
		goodsSku.Id = id
		o.Read(&goodsSku)

		temp["goods"] = goodsSku

		//商品的数量
		resp,err := conn.Do("hget","cart_"+ strconv.Itoa(user.Id),id)
		count,_:=redis.Int(resp,err)
		temp["count"] = count
		sumPrice := goodsSku.Price * count
		temp["sumPrice"] = sumPrice

		totalPrice += sumPrice
		totalCount += count

		goods = append(goods,temp)
	}
	this.Data["goods"] = goods


	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["transfer"] = 10
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goodsId"] = ids

	//返回数据
	this.TplName = "place_order.html"
}

//处理订单信息
func(this*OrderController)HandleOrderInfo(){
	//获取数据  单选框获取数据，用的是复选框的方法
	addId,err1 :=this.GetInt("addId")
	payId,err2 :=this.GetInt("payId")
	//js获取页面数据都是以字符串类型获取
	goodsId:=this.GetString("goodsId")
	totalPrice,err3 := this.GetInt("totalPrice")
	totalCount,err4 :=this.GetInt("totalCount")

	re := make(map[string]interface{})


	//校验数据
	if err1 != nil || err2 != nil || err3 != nil ||err4 != nil || len(goodsId) == 0{
		beego.Error("获取数据失败")
	}
	ids :=strings.Split(goodsId[1:len(goodsId)-1]," ")

	//处理数据
	//向订单表和订单商品表插入数据
	o := orm.NewOrm()
	var order models.OrderInfo
	order.TransitPrice = 10
	order.TotalPrice = totalPrice
	order.TotalCount = totalCount
	order.PayMethod = payId

	//获取用户数据
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user,"Name")

	order.OrderId = time.Now().Format("20060102150405")+strconv.Itoa(user.Id)
	order.User = &user

	//获取地址信息
	var addr models.Address
	addr.Id = addId
	o.Read(&addr)
	order.Address = &addr
	o.Begin()
	//插入操作
	o.Insert(&order)
	//插入数据到订单商品表
	conn,_ :=redis.Dial("tcp","192.168.110.81:6379")
	for _,value := range ids{
		id,_ :=strconv.Atoi(value)
		//获取商品信息
		for i:=0 ;i<3; i++{
			var goodsSku models.GoodsSKU
			goodsSku.Id = id
			o.Read(&goodsSku)
			//获取商品数量
			resp ,err :=conn.Do("hget","cart_"+strconv.Itoa(user.Id),id)
			count,_ :=redis.Int(resp,err)

			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goodsSku
			orderGoods.Price = goodsSku.Price * count
			orderGoods.OrderInfo = &order
			orderGoods.Count = count


			//插入操作
			o.Insert(&orderGoods)
			preStock := goodsSku.Stock
			if count > goodsSku.Stock{
				beego.Error("库存不足")
				re["code"] = 1
				re["errmsg"] = "商品库存不足，订单提交失败"
				this.Data["json"] = re
				o.Rollback()
				return
			}
			time.Sleep(time.Second * 10)


			//o.Update(&goodsSku)
			_,err =o.QueryTable("GoodsSKU").Filter("Id",goodsSku.Id).Filter("Stock",preStock).Update(orm.Params{"Stock":goodsSku.Stock-count,"Sales":goodsSku.Sales+count})
			if err != nil{
				o.Rollback()
				beego.Error("库存不足")
				re["code"] = 2
				re["errmsg"] = "商品库存不足，订单提交失败"
				this.Data["json"] = re
				o.Rollback()
				continue
			}else{
				break
			}
		}



		conn.Do("hdel","cart_"+strconv.Itoa(user.Id),id)

	}


	//返回数据
	re["code"] = 5
	re["errmsg"] = "OK"
	this.Data["json"] = re
	this.ServeJSON()
	o.Commit()
}
