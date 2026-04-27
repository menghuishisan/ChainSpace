package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

// CourseService 课程服务
type CourseService struct {
	courseRepo           *repository.CourseRepository
	courseStudentRepo    *repository.CourseStudentRepository
	chapterRepo          *repository.ChapterRepository
	materialRepo         *repository.MaterialRepository
	materialProgressRepo *repository.MaterialProgressRepository
	userRepo             *repository.UserRepository
	classRepo            *repository.ClassRepository
}

// NewCourseService 创建课程服务
func NewCourseService(
	courseRepo *repository.CourseRepository,
	courseStudentRepo *repository.CourseStudentRepository,
	chapterRepo *repository.ChapterRepository,
	materialRepo *repository.MaterialRepository,
	materialProgressRepo *repository.MaterialProgressRepository,
	userRepo *repository.UserRepository,
	classRepo *repository.ClassRepository,
) *CourseService {
	return &CourseService{
		courseRepo:           courseRepo,
		courseStudentRepo:    courseStudentRepo,
		chapterRepo:          chapterRepo,
		materialRepo:         materialRepo,
		materialProgressRepo: materialProgressRepo,
		userRepo:             userRepo,
		classRepo:            classRepo,
	}
}

// sameCourseSchool 判断课程是否属于当前学校。
func sameCourseSchool(resourceSchoolID uint, schoolID uint) bool {
	return schoolID > 0 && resourceSchoolID == schoolID
}

// canManageCourse 判断当前用户是否有权限管理指定课程。
func (s *CourseService) canManageCourse(course *model.Course, userID uint, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if !sameCourseSchool(course.SchoolID, schoolID) {
		return false
	}
	if role == model.RoleSchoolAdmin {
		return true
	}
	return role == model.RoleTeacher && course.TeacherID == userID
}

// canViewCourse 判断当前用户是否有权限查看指定课程。
func (s *CourseService) canViewCourse(ctx context.Context, course *model.Course, userID *uint, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if !sameCourseSchool(course.SchoolID, schoolID) {
		return false
	}
	if role == model.RoleSchoolAdmin || role == model.RoleTeacher {
		return true
	}
	if role == model.RoleStudent {
		if course.IsPublished() {
			return true
		}
		if userID == nil {
			return false
		}
		exists, err := s.courseStudentRepo.Exists(ctx, course.ID, *userID)
		return err == nil && exists
	}
	return false
}

func (s *CourseService) getAccessibleCourse(ctx context.Context, courseID uint, userID *uint, schoolID uint, role string) (*model.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canViewCourse(ctx, course, userID, schoolID, role) {
		return nil, errors.ErrCourseNotFound
	}
	return course, nil
}

// CreateCourse 创建课程
func (s *CourseService) CreateCourse(ctx context.Context, schoolID, teacherID uint, req *request.CreateCourseRequest) (*response.CourseResponse, error) {
	// 生成课程码
	code, err := s.courseRepo.GenerateCode(ctx)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 生成邀请码
	inviteCode, err := s.courseRepo.GenerateInviteCode(ctx)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	course := &model.Course{
		SchoolID:    schoolID,
		TeacherID:   teacherID,
		Title:       req.Title,
		Description: req.Description,
		Cover:       req.Cover,
		Code:        code,
		InviteCode:  inviteCode,
		Category:    req.Category,
		Tags:        req.Tags,
		Status:      "draft",
		IsPublic:    req.IsPublic,
		MaxStudents: req.MaxStudents,
	}

	// 解析日期
	if req.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始日期格式错误，需要RFC3339格式")
		}
		course.StartDate = &startDate
	}
	if req.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束日期格式错误，需要RFC3339格式")
		}
		course.EndDate = &endDate
	}

	if err := s.courseRepo.Create(ctx, course); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 获取完整信息
	course, _ = s.courseRepo.GetByID(ctx, course.ID)

	resp := &response.CourseResponse{}
	resp.FromCourse(course)
	return resp, nil
}

// UpdateCourse 更新课程
func (s *CourseService) UpdateCourse(ctx context.Context, courseID, userID, schoolID uint, role string, req *request.UpdateCourseRequest) (*response.CourseResponse, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Cover != "" {
		course.Cover = req.Cover
	}
	if req.Category != "" {
		course.Category = req.Category
	}
	if req.Tags != nil {
		course.Tags = req.Tags
	}
	if req.IsPublic != nil {
		course.IsPublic = *req.IsPublic
	}
	if req.MaxStudents != nil {
		course.MaxStudents = *req.MaxStudents
	}
	if req.Status != "" {
		course.Status = req.Status
	}
	if req.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("开始日期格式错误，需要RFC3339格式")
		}
		course.StartDate = &startDate
	}
	if req.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return nil, errors.ErrInvalidParams.WithMessage("结束日期格式错误，需要RFC3339格式")
		}
		course.EndDate = &endDate
	}

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.CourseResponse{}
	resp.FromCourse(course)
	return resp, nil
}

// DeleteCourse 删除课程
func (s *CourseService) DeleteCourse(ctx context.Context, courseID, userID, schoolID uint, role string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	if err := s.courseRepo.Delete(ctx, courseID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

// GetCourse 获取课程
func (s *CourseService) GetCourse(ctx context.Context, courseID uint, userID *uint, schoolID uint, role string) (*response.CourseDetailResponse, error) {
	course, err := s.courseRepo.GetWithChapters(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canViewCourse(ctx, course, userID, schoolID, role) {
		return nil, errors.ErrCourseNotFound
	}

	resp := &response.CourseDetailResponse{}
	resp.CourseResponse.FromCourse(course)

	// 统计学生数
	studentCount, _ := s.courseRepo.CountStudents(ctx, courseID)
	resp.StudentCount = studentCount

	// 统计章节数
	chapterCount, _ := s.courseRepo.CountChapters(ctx, courseID)
	resp.ChapterCount = chapterCount

	// 构建章节列表
	resp.Chapters = make([]response.ChapterResponse, len(course.Chapters))
	for i, chapter := range course.Chapters {
		chapterResp := &response.ChapterResponse{}
		chapterResp.FromChapter(&chapter)

		// 资料列表
		chapterResp.Materials = make([]response.MaterialResponse, len(chapter.Materials))
		for j, material := range chapter.Materials {
			materialResp := &response.MaterialResponse{}
			materialResp.FromMaterial(&material)

			// 获取学生学习进度
			if userID != nil {
				progress, err := s.materialProgressRepo.GetByID(ctx, material.ID, *userID)
				if err == nil {
					materialResp.Progress = progress.Progress
					materialResp.Completed = progress.Completed
				}
			}

			chapterResp.Materials[j] = *materialResp
		}

		resp.Chapters[i] = *chapterResp
	}

	// 获取学生课程进度
	if userID != nil {
		progress, err := s.materialProgressRepo.GetCourseProgress(ctx, courseID, *userID)
		if err == nil {
			resp.Progress = progress
		}
	}

	return resp, nil
}

// ListCourses 获取课程列表
func (s *CourseService) ListCourses(ctx context.Context, schoolID uint, req *request.ListCoursesRequest) ([]response.CourseResponse, int64, error) {
	courses, total, err := s.courseRepo.List(ctx, schoolID, req.TeacherID, req.Category, req.Status, req.Keyword, req.IsPublic, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.CourseResponse, len(courses))
	for i, course := range courses {
		resp := &response.CourseResponse{}
		resp.FromCourse(&course)

		// 统计学生数
		studentCount, _ := s.courseRepo.CountStudents(ctx, course.ID)
		resp.StudentCount = studentCount

		list[i] = *resp
	}

	return list, total, nil
}

// ListMyCourses 获取我的课程（学生）
func (s *CourseService) ListMyCourses(ctx context.Context, studentID uint, page, pageSize int) ([]response.CourseResponse, int64, error) {
	courses, total, err := s.courseRepo.ListByStudent(ctx, studentID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.CourseResponse, len(courses))
	for i, course := range courses {
		resp := &response.CourseResponse{}
		resp.FromCourse(&course)

		// 获取学习进度
		progress, _ := s.materialProgressRepo.GetCourseProgress(ctx, course.ID, studentID)
		resp.Progress = progress

		list[i] = *resp
	}

	return list, total, nil
}

// JoinCourse 加入课程
func (s *CourseService) JoinCourse(ctx context.Context, studentID uint, req *request.JoinCourseRequest) error {
	// 通过邀请码查找课程
	course, err := s.courseRepo.GetByInviteCode(ctx, req.Code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrInvalidCourseCode
		}
		return errors.ErrDatabaseError.WithError(err)
	}

	// 检查课程是否已发布
	if !course.IsPublished() {
		return errors.ErrCourseNotPublished
	}

	// 检查是否已加入
	exists, err := s.courseStudentRepo.Exists(ctx, course.ID, studentID)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if exists {
		return errors.ErrAlreadyEnrolled
	}

	// 检查人数限制
	if course.MaxStudents > 0 {
		count, _ := s.courseRepo.CountStudents(ctx, course.ID)
		if int(count) >= course.MaxStudents {
			return errors.ErrConflict.WithMessage("课程人数已满")
		}
	}

	// 创建关联
	cs := &model.CourseStudent{
		CourseID:  course.ID,
		StudentID: studentID,
		Status:    model.StatusActive,
	}

	if err := s.courseStudentRepo.Create(ctx, cs); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// AddStudentsToCourse 添加学生到课程
// 支持通过手机号或学号添加学生
func (s *CourseService) AddStudentsToCourse(ctx context.Context, courseID, userID, schoolID uint, role string, phones []string, studentNos []string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}

	var studentIDs []uint

	// 通过手机号查找
	for _, phone := range phones {
		if phone == "" {
			continue
		}
		user, err := s.userRepo.GetByPhone(ctx, phone)
		if err == nil && user != nil {
			studentIDs = append(studentIDs, user.ID)
		}
	}

	// 通过学号查找（学号全局唯一）
	for _, studentNo := range studentNos {
		if studentNo == "" {
			continue
		}
		user, err := s.userRepo.GetByStudentNo(ctx, 0, studentNo)
		if err == nil && user != nil {
			// 避免重复添加
			exists := false
			for _, id := range studentIDs {
				if id == user.ID {
					exists = true
					break
				}
			}
			if !exists {
				studentIDs = append(studentIDs, user.ID)
			}
		}
	}

	if len(studentIDs) == 0 {
		return errors.ErrInvalidParams.WithError(fmt.Errorf("未找到任何有效学生"))
	}

	for _, studentID := range studentIDs {
		exists, err := s.courseStudentRepo.Exists(ctx, courseID, studentID)
		if err != nil {
			return errors.ErrDatabaseError.WithError(err)
		}
		if exists {
			continue
		}

		cs := &model.CourseStudent{
			CourseID:  courseID,
			StudentID: studentID,
			Status:    model.StatusActive,
		}

		if err := s.courseStudentRepo.Create(ctx, cs); err != nil {
			return errors.ErrDatabaseError.WithError(err)
		}
	}

	return nil
}

// RemoveStudentsFromCourse 从课程移除学生
func (s *CourseService) RemoveStudentsFromCourse(ctx context.Context, courseID, userID, schoolID uint, role string, studentIDs []uint) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}

	if err := s.courseStudentRepo.BatchDelete(ctx, courseID, studentIDs); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

// ListCourseStudents 获取课程学生列表
func (s *CourseService) ListCourseStudents(ctx context.Context, courseID, userID, schoolID uint, role string, req *request.ListCourseStudentsRequest) ([]response.CourseStudentResponse, int64, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, errors.ErrCourseNotFound
		}
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, 0, errors.ErrNoPermission
	}

	list, total, err := s.courseStudentRepo.List(ctx, courseID, req.Status, req.Keyword, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	result := make([]response.CourseStudentResponse, len(list))
	for i, cs := range list {
		result[i] = response.CourseStudentResponse{
			ID:         cs.ID,
			StudentID:  cs.StudentID,
			Progress:   cs.Progress,
			LastAccess: cs.LastAccess,
			JoinedAt:   cs.JoinedAt,
			Status:     cs.Status,
		}
		if cs.Student != nil {
			result[i].RealName = cs.Student.RealName
			result[i].StudentNo = cs.Student.StudentNo
			if cs.Student.Class != nil {
				result[i].ClassName = cs.Student.Class.Name
			}
		}
	}

	return result, total, nil
}

// ImportStudentsFromExcel 从Excel导入学生到课程
func (s *CourseService) ImportStudentsFromExcel(ctx context.Context, courseID, userID, schoolID uint, role string, file *multipart.FileHeader) (*ImportResult, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.ErrInvalidParams.WithError(err)
	}
	defer src.Close()

	// 读取Excel文件
	records, err := readExcelRecords(src)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithError(err)
	}

	result := &ImportResult{
		Success: 0,
		Failed:  0,
		Errors:  []string{},
	}

	// 获取课程信息以获取学校ID
	for i, record := range records {
		rowNum := i + 2 // Excel行号从2开始（1是表头）

		// 必填字段验证
		phone := getStringField(record, "phone")
		studentNo := getStringField(record, "student_no")
		realName := getStringField(record, "real_name")
		className := getStringField(record, "class_name")

		if phone == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 手机号为空", rowNum))
			result.Failed++
			continue
		}
		if studentNo == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 学号为空", rowNum))
			result.Failed++
			continue
		}
		if realName == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 真实姓名为空", rowNum))
			result.Failed++
			continue
		}
		if className == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 班级为空", rowNum))
			result.Failed++
			continue
		}

		// 查找或创建用户（按手机号查找）
		user, err := s.userRepo.GetByPhone(ctx, phone)
		if err != nil {
			// 用户不存在，创建新用户
			user = &model.User{
				Phone:         phone,
				RealName:      realName,
				StudentNo:     studentNo,
				Role:          model.RoleStudent,
				Status:        model.StatusActive,
				MustChangePwd: true,
			}
			// 设置学校ID
			if course.SchoolID > 0 {
				user.SchoolID = &course.SchoolID

				// 查找或创建班级
				class, err := s.classRepo.GetByName(ctx, course.SchoolID, className)
				if err != nil {
					// 班级不存在，创建班级
					class = &model.Class{
						SchoolID: course.SchoolID,
						Name:     className,
						Status:   model.StatusActive,
					}
					if err := s.classRepo.Create(ctx, class); err == nil {
						user.ClassID = &class.ID
					}
				} else {
					user.ClassID = &class.ID
				}
			}

			if err := s.userRepo.Create(ctx, user); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 创建用户失败 - %s", rowNum, err.Error()))
				result.Failed++
				continue
			}
			// 重新获取用户（包含ID）
			user, _ = s.userRepo.GetByPhone(ctx, phone)
		}

		// 检查是否已在课程中
		exists, _ := s.courseStudentRepo.Exists(ctx, courseID, user.ID)
		if exists {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 用户 %s 已在课程中", rowNum, user.RealName))
			result.Failed++
			continue
		}

		// 添加到课程
		cs := &model.CourseStudent{
			CourseID:  courseID,
			StudentID: user.ID,
			Status:    model.StatusActive,
		}
		if err := s.courseStudentRepo.Create(ctx, cs); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 添加到课程失败 - %s", rowNum, err.Error()))
			result.Failed++
			continue
		}

		result.Success++
	}

	return result, nil
}

// ImportResult 导入结果
type ImportResult struct {
	Success int
	Failed  int
	Errors  []string
}

// readExcelRecords 读取Excel记录
func readExcelRecords(reader io.Reader) ([]map[string]string, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("文件内容为空或只有表头")
	}

	// 解析表头
	headers := records[0]
	result := make([]map[string]string, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		row := records[i]
		record := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				record[strings.TrimSpace(strings.ToLower(header))] = strings.TrimSpace(row[j])
			}
		}
		result = append(result, record)
	}

	return result, nil
}

// getStringField 获取字符串字段
func getStringField(record map[string]string, field string) string {
	return record[strings.ToLower(field)]
}

// CreateChapter 创建章节
func (s *CourseService) CreateChapter(ctx context.Context, courseID, userID, schoolID uint, role string, req *request.CreateChapterRequest) (*response.ChapterResponse, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	// 获取最大排序号
	maxOrder, _ := s.chapterRepo.GetMaxSortOrder(ctx, courseID)

	chapter := &model.Chapter{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		SortOrder:   maxOrder + 1,
		Status:      "draft",
	}

	if req.SortOrder > 0 {
		chapter.SortOrder = req.SortOrder
	}

	if err := s.chapterRepo.Create(ctx, chapter); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ChapterResponse{}
	resp.FromChapter(chapter)
	return resp, nil
}

// UpdateChapter 更新章节
func (s *CourseService) UpdateChapter(ctx context.Context, chapterID, userID, schoolID uint, role string, req *request.UpdateChapterRequest) (*response.ChapterResponse, error) {
	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrChapterNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	course, err := s.courseRepo.GetByID(ctx, chapter.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	if req.Title != "" {
		chapter.Title = req.Title
	}
	if req.Description != "" {
		chapter.Description = req.Description
	}
	if req.SortOrder != nil {
		chapter.SortOrder = *req.SortOrder
	}
	if req.Status != "" {
		chapter.Status = req.Status
	}

	if err := s.chapterRepo.Update(ctx, chapter); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.ChapterResponse{}
	resp.FromChapter(chapter)
	return resp, nil
}

// DeleteChapter 删除章节
func (s *CourseService) DeleteChapter(ctx context.Context, chapterID, userID, schoolID uint, role string) error {
	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrChapterNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	course, err := s.courseRepo.GetByID(ctx, chapter.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	if err := s.chapterRepo.Delete(ctx, chapterID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

// ListChapters 获取课程章节列表
func (s *CourseService) ListChapters(ctx context.Context, courseID uint, userID *uint, schoolID uint, role string) ([]response.ChapterResponse, error) {
	if _, err := s.getAccessibleCourse(ctx, courseID, userID, schoolID, role); err != nil {
		return nil, err
	}

	chapters, err := s.chapterRepo.ListByCourse(ctx, courseID, "")
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ChapterResponse, 0, len(chapters))
	for _, ch := range chapters {
		r := response.ChapterResponse{}
		r.FromChapter(&ch)
		list = append(list, r)
	}
	return list, nil
}

// ListMaterials 获取章节资料列表
func (s *CourseService) ListMaterials(ctx context.Context, courseID uint, chapterID uint, userID *uint, schoolID uint, role string) ([]response.MaterialResponse, error) {
	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrChapterNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if chapter.CourseID != courseID {
		return nil, errors.ErrChapterNotFound
	}
	if _, err := s.getAccessibleCourse(ctx, courseID, userID, schoolID, role); err != nil {
		return nil, err
	}

	materials, err := s.materialRepo.ListByChapter(ctx, chapterID, "", "")
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.MaterialResponse, 0, len(materials))
	for _, m := range materials {
		r := response.MaterialResponse{}
		r.FromMaterial(&m)
		list = append(list, r)
	}
	return list, nil
}

// ReorderChapters 重新排序章节
func (s *CourseService) ReorderChapters(ctx context.Context, courseID, userID, schoolID uint, role string, chapterIDs []uint) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}

	orders := make(map[uint]int)
	for i, id := range chapterIDs {
		orders[id] = i + 1
	}

	if err := s.chapterRepo.BatchUpdateSortOrder(ctx, orders); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}

	return nil
}

// CreateMaterial 创建资料
func (s *CourseService) CreateMaterial(ctx context.Context, chapterID, userID, schoolID uint, role string, req *request.CreateMaterialRequest) (*response.MaterialResponse, error) {
	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrChapterNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	course, err := s.courseRepo.GetByID(ctx, chapter.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	// 获取最大排序号
	maxOrder, _ := s.materialRepo.GetMaxSortOrder(ctx, chapterID)

	material := &model.Material{
		ChapterID: chapterID,
		Title:     req.Title,
		Type:      req.Type,
		Content:   req.Content,
		URL:       req.URL,
		Duration:  req.Duration,
		SortOrder: maxOrder + 1,
		Status:    model.StatusActive,
	}

	if req.SortOrder > 0 {
		material.SortOrder = req.SortOrder
	}

	if err := s.materialRepo.Create(ctx, material); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.MaterialResponse{}
	resp.FromMaterial(material)
	return resp, nil
}

// UpdateMaterial 更新资料
func (s *CourseService) UpdateMaterial(ctx context.Context, materialID, userID, schoolID uint, role string, req *request.UpdateMaterialRequest) (*response.MaterialResponse, error) {
	material, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrMaterialNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	chapter, err := s.chapterRepo.GetByID(ctx, material.ChapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrChapterNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	course, err := s.courseRepo.GetByID(ctx, chapter.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCourseNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return nil, errors.ErrNoPermission
	}

	if req.Title != "" {
		material.Title = req.Title
	}
	if req.Type != "" {
		material.Type = req.Type
	}
	if req.Content != "" {
		material.Content = req.Content
	}
	if req.URL != "" {
		material.URL = req.URL
	}
	if req.Duration != nil {
		material.Duration = *req.Duration
	}
	if req.SortOrder != nil {
		material.SortOrder = *req.SortOrder
	}
	if req.Status != "" {
		material.Status = req.Status
	}

	if err := s.materialRepo.Update(ctx, material); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	resp := &response.MaterialResponse{}
	resp.FromMaterial(material)
	return resp, nil
}

// DeleteMaterial 删除资料
func (s *CourseService) DeleteMaterial(ctx context.Context, materialID, userID, schoolID uint, role string) error {
	material, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrMaterialNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	chapter, err := s.chapterRepo.GetByID(ctx, material.ChapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrChapterNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	course, err := s.courseRepo.GetByID(ctx, chapter.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCourseNotFound
		}
		return errors.ErrDatabaseError.WithError(err)
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	if err := s.materialRepo.Delete(ctx, materialID); err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	return nil
}

// UpdateMaterialProgress 更新学习进度
func (s *CourseService) UpdateMaterialProgress(ctx context.Context, courseID, materialID, studentID, schoolID uint, role string, req *request.UpdateMaterialProgressRequest) (*response.MaterialProgressResponse, error) {
	if role != model.RoleStudent {
		return nil, errors.ErrNoPermission
	}

	material, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrMaterialNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	chapter, err := s.chapterRepo.GetByID(ctx, material.ChapterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrChapterNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if chapter.CourseID != courseID {
		return nil, errors.ErrMaterialNotFound
	}
	if _, err := s.getAccessibleCourse(ctx, courseID, &studentID, schoolID, role); err != nil {
		return nil, err
	}

	progress, err := s.materialProgressRepo.GetByID(ctx, materialID, studentID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if progress == nil {
		progress = &model.MaterialProgress{
			MaterialID: materialID,
			StudentID:  studentID,
		}
	}

	progress.Progress = req.Progress
	progress.Duration = req.Duration
	progress.LastPosition = req.LastPosition

	if req.Progress >= 100 {
		progress.Completed = true
		now := time.Now()
		progress.CompletedAt = &now
	}

	if err := s.materialProgressRepo.Upsert(ctx, progress); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	// 更新课程学习进度
	courseProgress, _ := s.materialProgressRepo.GetCourseProgress(ctx, courseID, studentID)
	_ = s.courseStudentRepo.UpdateProgress(ctx, courseID, studentID, courseProgress)

	return &response.MaterialProgressResponse{
		MaterialID:   progress.MaterialID,
		Progress:     progress.Progress,
		Duration:     progress.Duration,
		LastPosition: progress.LastPosition,
		Completed:    progress.Completed,
		CompletedAt:  progress.CompletedAt,
	}, nil
}

// ResetInviteCode 重置课程邀请码
func (s *CourseService) ResetInviteCode(ctx context.Context, courseID, userID, schoolID uint, role string) (string, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return "", errors.ErrCourseNotFound
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return "", errors.ErrNoPermission
	}

	newCode, err := s.courseRepo.GenerateInviteCode(ctx)
	if err != nil {
		return "", errors.ErrDatabaseError.WithError(err)
	}

	course.InviteCode = newCode
	if err := s.courseRepo.Update(ctx, course); err != nil {
		return "", errors.ErrDatabaseError.WithError(err)
	}

	return newCode, nil
}

// UpdateCourseStatus 更新课程状态
func (s *CourseService) UpdateCourseStatus(ctx context.Context, courseID, userID, schoolID uint, role string, status string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.ErrCourseNotFound
	}
	if !s.canManageCourse(course, userID, schoolID, role) {
		return errors.ErrNoPermission
	}

	course.Status = status
	return s.courseRepo.Update(ctx, course)
}

// GetCourseProgress 获取课程学习进度
func (s *CourseService) GetCourseProgress(ctx context.Context, courseID, studentID, schoolID uint, role string) (*response.CourseProgressResponse, error) {
	if _, err := s.getAccessibleCourse(ctx, courseID, &studentID, schoolID, role); err != nil {
		return nil, err
	}

	// 获取课程所有资料数量
	totalMaterials, _ := s.materialRepo.CountByCourse(ctx, courseID)
	// 获取已完成的资料数量
	completedMaterials, _ := s.materialProgressRepo.CountCompletedByCourse(ctx, courseID, studentID)

	// 获取课程所有实验数量
	totalExperiments, _ := s.courseRepo.CountExperiments(ctx, courseID)
	// 获取已完成的实验数量
	completedExperiments, _ := s.courseRepo.CountCompletedExperiments(ctx, courseID, studentID)

	var progressPercent int
	total := totalMaterials + totalExperiments
	if total > 0 {
		completed := completedMaterials + completedExperiments
		progressPercent = int(float64(completed) / float64(total) * 100)
	}

	return &response.CourseProgressResponse{
		CourseID:             courseID,
		TotalMaterials:       int(totalMaterials),
		CompletedMaterials:   int(completedMaterials),
		TotalExperiments:     int(totalExperiments),
		CompletedExperiments: int(completedExperiments),
		ProgressPercent:      progressPercent,
	}, nil
}
