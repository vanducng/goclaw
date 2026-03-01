package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
)

// FriendInfo is a minimal friend/contact record for the picker UI.
type FriendInfo struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	ZaloName    string `json:"zaloName,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// GroupListInfo is a minimal group record for the picker UI.
type GroupListInfo struct {
	GroupID     string `json:"groupId"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar,omitempty"`
	TotalMember int    `json:"totalMember"`
}

// FetchFriends fetches the authenticated user's friend list from Zalo.
func FetchFriends(ctx context.Context, sess *Session) ([]FriendInfo, error) {
	baseURL := getServiceURL(sess, "profile")
	if baseURL == "" {
		return nil, fmt.Errorf("zca: no profile service URL")
	}

	payload := map[string]any{
		"page":        1,
		"count":       20000,
		"incInvalid":  1,
		"avatar_size": 120,
		"actiontime":  0,
		"imei":        sess.IMEI,
	}

	encData, err := encryptPayload(sess, payload)
	if err != nil {
		return nil, fmt.Errorf("zca: encrypt friends payload: %w", err)
	}

	reqURL := makeURL(sess, baseURL+"/api/social/friend/getfriends",
		map[string]any{"params": encData}, true)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sess)

	resp, err := sess.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zca: fetch friends: %w", err)
	}
	defer resp.Body.Close()

	// Response envelope: {"error_code":0, "data":"<encrypted_base64>"}
	var envelope Response[*string]
	if err := readJSON(resp, &envelope); err != nil {
		return nil, fmt.Errorf("zca: parse friends response: %w", err)
	}
	if envelope.ErrorCode != 0 {
		return nil, fmt.Errorf("zca: friends error code %d: %s", envelope.ErrorCode, envelope.ErrorMessage)
	}
	if envelope.Data == nil {
		return nil, fmt.Errorf("zca: empty friends data")
	}

	plain, err := decryptDataField(sess, *envelope.Data)
	if err != nil {
		return nil, fmt.Errorf("zca: decrypt friends: %w", err)
	}

	var friends []FriendInfo
	if err := json.Unmarshal(plain, &friends); err != nil {
		return nil, fmt.Errorf("zca: parse friends list: %w", err)
	}
	return friends, nil
}

// FetchGroups fetches the authenticated user's group list from Zalo (two-step).
func FetchGroups(ctx context.Context, sess *Session) ([]GroupListInfo, error) {
	// Step 1: Get group IDs from group_poll service
	gridVerMap, err := fetchGroupIDs(ctx, sess)
	if err != nil {
		return nil, err
	}
	if len(gridVerMap) == 0 {
		return nil, nil
	}

	// Step 2: Get group details from group service
	return fetchGroupDetails(ctx, sess, gridVerMap)
}

// fetchGroupIDs gets group ID -> version map from group_poll service.
func fetchGroupIDs(ctx context.Context, sess *Session) (map[string]string, error) {
	baseURL := getServiceURL(sess, "group_poll")
	if baseURL == "" {
		return nil, fmt.Errorf("zca: no group_poll service URL")
	}

	reqURL := makeURL(sess, baseURL+"/api/group/getlg/v4", nil, true)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sess)

	resp, err := sess.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zca: fetch group IDs: %w", err)
	}
	defer resp.Body.Close()

	var envelope Response[*string]
	if err := readJSON(resp, &envelope); err != nil {
		return nil, fmt.Errorf("zca: parse group IDs response: %w", err)
	}
	if envelope.ErrorCode != 0 {
		return nil, fmt.Errorf("zca: group IDs error code %d: %s", envelope.ErrorCode, envelope.ErrorMessage)
	}
	if envelope.Data == nil {
		return nil, nil
	}

	plain, err := decryptDataField(sess, *envelope.Data)
	if err != nil {
		return nil, fmt.Errorf("zca: decrypt group IDs: %w", err)
	}

	var result struct {
		GridVerMap map[string]string `json:"gridVerMap"`
	}
	if err := json.Unmarshal(plain, &result); err != nil {
		return nil, fmt.Errorf("zca: parse group IDs: %w", err)
	}
	return result.GridVerMap, nil
}

// fetchGroupDetails gets group info for given group IDs.
func fetchGroupDetails(ctx context.Context, sess *Session, gridVerMap map[string]string) ([]GroupListInfo, error) {
	baseURL := getServiceURL(sess, "group")
	if baseURL == "" {
		return nil, fmt.Errorf("zca: no group service URL")
	}

	// Build payload with gridVerMap as JSON string
	gridVerJSON, err := json.Marshal(gridVerMap)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"gridVerMap": string(gridVerJSON),
	}

	encData, err := encryptPayload(sess, payload)
	if err != nil {
		return nil, fmt.Errorf("zca: encrypt group details payload: %w", err)
	}

	reqURL := makeURL(sess, baseURL+"/api/group/getmg-v2", nil, true)

	form := buildFormBody(map[string]string{"params": encData})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, form)
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sess)

	resp, err := sess.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zca: fetch group details: %w", err)
	}
	defer resp.Body.Close()

	var envelope Response[*string]
	if err := readJSON(resp, &envelope); err != nil {
		return nil, fmt.Errorf("zca: parse group details response: %w", err)
	}
	if envelope.ErrorCode != 0 {
		return nil, fmt.Errorf("zca: group details error code %d: %s", envelope.ErrorCode, envelope.ErrorMessage)
	}
	if envelope.Data == nil {
		return nil, nil
	}

	plain, err := decryptDataField(sess, *envelope.Data)
	if err != nil {
		return nil, fmt.Errorf("zca: decrypt group details: %w", err)
	}

	var result struct {
		GridInfoMap map[string]struct {
			Name        string `json:"name"`
			Avatar      string `json:"avt"`
			TotalMember int    `json:"totalMember"`
		} `json:"gridInfoMap"`
	}
	if err := json.Unmarshal(plain, &result); err != nil {
		return nil, fmt.Errorf("zca: parse group details: %w", err)
	}

	groups := make([]GroupListInfo, 0, len(result.GridInfoMap))
	for id, info := range result.GridInfoMap {
		groups = append(groups, GroupListInfo{
			GroupID:     id,
			Name:        info.Name,
			Avatar:      info.Avatar,
			TotalMember: info.TotalMember,
		})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
	return groups, nil
}

// decryptDataField decrypts an encrypted base64 data string from Zalo API response.
func decryptDataField(sess *Session, data string) ([]byte, error) {
	key := SecretKey(sess.SecretKey).Bytes()
	if key == nil {
		return nil, fmt.Errorf("zca: invalid session secret key")
	}
	unescaped, err := url.PathUnescape(data)
	if err != nil {
		return nil, err
	}
	return DecodeAESCBC(key, unescaped)
}
