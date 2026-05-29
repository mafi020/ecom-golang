package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/config"
	"github.com/mafi020/ecom-golang/internal/delivery/http/handler"
	"github.com/mafi020/ecom-golang/internal/delivery/http/middleware"
)

// Per route rate limiter example - 10 requests per minute on auth
// authLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/10), 5)

// auth := r.Group("/auth")
// auth.Use(authLimiter.Middleware())
// {
//     auth.POST("/login",    authHandler.Login)
//     auth.POST("/register", authHandler.Register)
// }

func RegisterHTTPHandlers(r *gin.Engine, uc *Usecases, cfg *config.Config) {
	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	authHandler := handler.NewAuthHandler(uc.AuthUC)
	userHandler := handler.NewUserHandler(uc.UserUC)
	categoryHandler := handler.NewCategoryHandler(uc.CategoryUC)
	productHandler := handler.NewProductHandler(uc.ProductUC)
	orderHandler := handler.NewOrderHandler(uc.OrderUC)
	cartHandler := handler.NewCartHandler(uc.CartUC)
	paymentHandler := handler.NewPaymentHandler(uc.PaymentUC)

	api := r.Group("/api")
	{
		// Public routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/:id", userHandler.GetUserByID)
				userGroup.GET("/", userHandler.GetUsers)
			}

			categoryGroup := protected.Group("/categories")
			{
				categoryGroup.POST("/", categoryHandler.CreateCategory)
				categoryGroup.GET("/:id", categoryHandler.GetCategoryByID)
				categoryGroup.GET("/:id/products", categoryHandler.GetCategoryByIDWithProducts)
				categoryGroup.GET("/", categoryHandler.GetAllCategories)
				categoryGroup.PUT("/:id", categoryHandler.UpdateCategory)
				categoryGroup.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			productGroup := protected.Group("/products")
			{
				productGroup.POST("/", productHandler.CreateProduct)
				productGroup.GET("/:id", productHandler.GetProductByID)
				productGroup.GET("/", productHandler.GetProducts)
				productGroup.PUT("/:id", productHandler.UpdateProduct)
				productGroup.DELETE("/:id", productHandler.DeleteProduct)
			}

			cartGroup := protected.Group("/cart")
			{
				cartGroup.GET("", cartHandler.GetCart)
				cartGroup.DELETE("", cartHandler.ClearCart)
				cartGroup.POST("/items", cartHandler.AddItem)
				cartGroup.PUT("/items/:product_id", cartHandler.UpdateItem)
				cartGroup.DELETE("/items/:product_id", cartHandler.RemoveItem)
			}

			orderGroup := protected.Group("/orders")
			{
				orderGroup.POST("/", orderHandler.PlaceOrder)
				orderGroup.GET("", orderHandler.GetOrdersByUserID) // /orders → my orders
				orderGroup.GET("/:id", orderHandler.GetOrderByID)  // /orders/:id → my order
			}

			// User routes
			paymentGroup := r.Group("/payments")
			{
				paymentGroup.POST("/online", paymentHandler.PayOnline)
				paymentGroup.POST("/cod", paymentHandler.PayCOD)
				paymentGroup.GET("/order/:order_id", paymentHandler.GetPaymentByOrderID)
			}

			// Admin routes
			// adminGroup := r.Group("/admin")
			// adminGroup.Use(middleware.Auth(cfg), middleware.RequireRole("admin"))
			// {
			// 	adminGroup.PUT("/payments/order/:order_id/collect", paymentHandler.CollectCOD)
			// 	adminGroup.PUT("/orders/:id/status", orderHandler.UpdateStatus)
			// }

		}
	}

}
