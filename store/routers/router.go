package routers

import (
	"bj2qFresh/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec,filterFunc)
	//首页业务
    beego.Router("/", &controllers.GoodsController{},"get:ShowIndex")
    //注册业务
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    beego.Router("/active",&controllers.UserController{},"get:HandleActive")
    //登录业务
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
    //退出登录
    beego.Router("/user/logout",&controllers.UserController{},"get:Logout")
    //用户中心信息页
    beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")
    //用户中心订单页
    beego.Router("/user/userCenterOrder",&controllers.UserController{},"get:ShowUserCenterOrder")
    //用户中心地址页
    beego.Router("/user/userCenterSite",&controllers.UserController{},"get:ShowUserCenterSite;post:HandleUserCenterSite")
    //展示文章详情页
    beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowGoodsDetail")
    //展示商品列表页
    beego.Router("/goodsList",&controllers.GoodsController{},"get:ShowGoodsList")
    //搜索商品
    beego.Router("/searchGoods",&controllers.GoodsController{},"post:HandleSearch")
    //添加购物车请求
    beego.Router("/user/addCart",&controllers.CartController{},"post:HandleAddCart")
    //展示购物车
    beego.Router("/user/showCart",&controllers.CartController{},"get:ShowCart")
    //更新购物车数量
    beego.Router("/user/updateCart",&controllers.CartController{},"post:UpdateCart")
    //删除购物车数据
    beego.Router("/user/deleteCart",&controllers.CartController{},"post:DeleteCart")
    //展示订单页
    beego.Router("/user/showOrder",&controllers.OrderController{},"post:ShowOrder")
    //提交订单
    beego.Router("/user/orderInfo",&controllers.OrderController{},"post:HandleOrderInfo")
}

var filterFunc = func(ctx*context.Context) {
	userName := ctx.Input.Session("userName")
	if userName == nil{
		ctx.Redirect(302,"/login")
	}
}
