package handles

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type PlayCountReportReq struct {
	Path string `json:"path" binding:"required"`
}

func ReportPlay(c *gin.Context) {
	var req PlayCountReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	err = op.ReportPlay(user.ID, reqPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "success")
}

type PlayCountQueryReq struct {
	Paths []string `json:"paths" binding:"required"`
}

type PlayCountQueryRespItem struct {
	Path       string `json:"path"`
	Count      int    `json:"count"`
	LastPlayed string `json:"last_played"`
}

func GetPlayCounts(c *gin.Context) {
	var req PlayCountQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)

	var queryPaths []string
	for _, p := range req.Paths {
		reqPath, err := user.JoinPath(p)
		if err == nil {
			queryPaths = append(queryPaths, reqPath)
		}
	}

	playCounts, err := op.GetPlayCounts(user.ID, queryPaths)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	var resp []PlayCountQueryRespItem
	for _, p := range req.Paths {
		reqPath, err := user.JoinPath(p)
		if err != nil {
			continue
		}
		count := 0
		lastPlayed := ""
		if pc, ok := playCounts[reqPath]; ok {
			count = pc.Count
			lastPlayed = pc.LastPlayed.Format("2006-01-02 15:04:05")
		}
		resp = append(resp, PlayCountQueryRespItem{
			Path:       p,
			Count:      count,
			LastPlayed: lastPlayed,
		})
	}

	common.SuccessResp(c, resp)
}
