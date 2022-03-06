package user

import (
	"fmt"
	"log"
	"net/http"

	"example.com/application/user"
	"example.com/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	FindByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type userHandler struct {
	user.UserService
}

func NewUserHandler(svc user.UserService) UserHandler {
	return &userHandler{svc}
}

func (uh *userHandler) FindByID(c *gin.Context) {
	var idReq userIDRequest
	if err := c.BindUri(&idReq); err != nil {
		return
	}

	ret, err := uh.UserService.FindByID(idReq.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ret)
	return
}

func (uh *userHandler) Create(c *gin.Context) {
	var req userRequest
	if err := c.Bind(&req); err != nil {
		return
	}
	currentTime := utils.CurrentTime()
	ud := &user.UserData{
		Name:      req.Name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	u, err := uh.UserService.Create(ud)
	if err != nil {
		log.Printf("%#v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Location", fmt.Sprintf("%s/%d", c.Request.URL.Path, u.ID))
	c.Status(http.StatusCreated)
	return
}

func (uh *userHandler) Update(c *gin.Context) {
	var idReq userIDRequest
	var req userRequest
	if err := c.BindUri(&idReq); err != nil {
		return
	}

	if err := c.Bind(&req); err != nil {
		return
	}

	ud := &user.UserData{
		ID:        idReq.ID,
		Name:      req.Name,
		UpdatedAt: utils.CurrentTime(),
	}

	if err := uh.UserService.Update(ud); err != nil {
		log.Printf("%#v\n", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
	return
}

func (uh *userHandler) Delete(c *gin.Context) {
	var idReq userIDRequest
	if err := c.BindUri(&idReq); err != nil {
		return
	}

	if err := uh.UserService.Delete(idReq.ID); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, "user is not exists")
		return
	}
	c.Status(http.StatusNoContent)
	return
}
