package svc

import (
	"context"
	"net/http"
	"sync"
)

// UserContextKey 用户上下文键
type UserContextKey struct{}

// UserInfo 用户信息结构
type UserInfo struct {
	UserID   int
	Username string
	RoleId   int
}

var (
	currentUser struct {
		UserID   int
		Username string
		RoleId   int
		mutex    sync.Mutex
	}
)

// SetCurrentUser 设置当前登录用户
func SetCurrentUser(userID int, username string, roleId int) {
	currentUser.mutex.Lock()
	defer currentUser.mutex.Unlock()
	currentUser.UserID = userID
	currentUser.Username = username
	currentUser.RoleId = roleId
}

// GetCurrentUser 获取当前登录用户
func GetCurrentUser() (int, string, int) {
	currentUser.mutex.Lock()
	defer currentUser.mutex.Unlock()
	return currentUser.UserID, currentUser.Username, currentUser.RoleId
}

// ... existing code ...
func SingleUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 直接从全局存储获取当前用户
		userID, username, roleId := GetCurrentUser()
		if userID == 0 {
			http.Error(w, "未登录", http.StatusUnauthorized)
			return
		}

		// 将用户信息存入Context
		ctx := context.WithValue(r.Context(), UserContextKey{}, UserInfo{
			UserID:   userID,
			Username: username,
			RoleId:   roleId,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// // 登录
// func (p loginNotifier) Handle(mgr *SrvMgr, payload ReqLogin) {
//     // Do something for this event
//     l.Log.Debug("Handle loginNotifier called")
//     res, err := mSrvMgr.sysUserPd.Login(payload.UserName, payload.Password)
//     if err != nil {
//         l.Log.Error(err)
//         mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_LOGIN, MsgBody: "fail,login failed"}
//         return
//     }
//     if res {
//         // 获取用户详细信息（需要补充实现sysUserPd.GetUserInfo方法）
//         userID, roleId, err := mSrvMgr.sysUserPd.GetUserInfo(payload.UserName)
//         if err != nil {
//             l.Log.Error(err)
//             mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_LOGIN, MsgBody: "fail,get user info failed"}
//             return
//         }
//         // 初始化当前用户
//         SetCurrentUser(userID, payload.UserName, roleId)
//         mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_LOGIN, MsgBody: "ok"}
//     } else {
//         mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_LOGIN, MsgBody: "fail,login failed"}
//     }
// }
