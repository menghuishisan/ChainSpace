/**
 * 路由配置
 */
import { lazy } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useUserStore } from '@/store'
import { RoleDefaultRoute } from '@/types'
import AuthGuard from './guards'

// ====== 懒加载页面组件 ======

// 认证页面
const Login = lazy(() => import('@/pages/auth/Login'))

// 布局
const MainLayout = lazy(() => import('@/components/layout/MainLayout'))

// 平台管理员页面
const AdminSchools = lazy(() => import('@/pages/admin/Schools'))
const AdminImages = lazy(() => import('@/pages/admin/Images'))
const AdminConfigs = lazy(() => import('@/pages/admin/Configs'))
const AdminStatistics = lazy(() => import('@/pages/admin/Statistics'))
const AdminContests = lazy(() => import('@/pages/admin/Contests'))
const AdminChallenges = lazy(() => import('@/pages/admin/Challenges'))
const AdminReviews = lazy(() => import('@/pages/admin/Reviews'))
const AdminVulnerabilities = lazy(() => import('@/pages/admin/Vulnerabilities'))
const AdminMonitor = lazy(() => import('@/pages/admin/Monitor'))

// 学校管理员页面
const SchoolTeachers = lazy(() => import('@/pages/school/Teachers'))
const SchoolStudents = lazy(() => import('@/pages/school/Students'))
const SchoolClasses = lazy(() => import('@/pages/school/Classes'))
const SchoolStatistics = lazy(() => import('@/pages/school/Statistics'))
const SchoolContests = lazy(() => import('@/pages/school/Contests'))
const SchoolSettings = lazy(() => import('@/pages/school/Settings'))
const SchoolChallenges = lazy(() => import('@/pages/school/Challenges'))
const SchoolCrossSchoolContests = lazy(() => import('@/pages/school/CrossSchoolContests'))

// 教师页面
const TeacherCourses = lazy(() => import('@/pages/teacher/Courses'))
const TeacherCourseDetail = lazy(() => import('@/pages/teacher/CourseDetail'))
const TeacherCourseCreate = lazy(() => import('@/pages/teacher/CourseCreate'))
const TeacherExperiments = lazy(() => import('@/pages/teacher/Experiments'))
const TeacherExperimentCreate = lazy(() => import('@/pages/teacher/ExperimentCreate'))
const TeacherGrading = lazy(() => import('@/pages/teacher/Grading'))
const TeacherStatistics = lazy(() => import('@/pages/teacher/Statistics'))
const TeacherContests = lazy(() => import('@/pages/teacher/Contests'))
const TeacherChallenges = lazy(() => import('@/pages/teacher/Challenges'))
const TeacherExperimentEdit = lazy(() => import('@/pages/teacher/ExperimentEdit'))
const TeacherDiscussionManage = lazy(() => import('@/pages/teacher/DiscussionManage'))
const TeacherContestMonitor = lazy(() => import('@/pages/teacher/ContestMonitor'))

// 学生页面
const StudentCourses = lazy(() => import('@/pages/student/Courses'))
const StudentCourseDetail = lazy(() => import('@/pages/student/CourseDetail'))
const StudentExperiments = lazy(() => import('@/pages/student/Experiments'))
const StudentExperimentDetail = lazy(() => import('@/pages/student/ExperimentDetail'))
const StudentExperimentEnv = lazy(() => import('@/pages/student/ExperimentEnv'))
const StudentGrades = lazy(() => import('@/pages/student/Grades'))
const StudentContests = lazy(() => import('@/pages/student/Contests'))
const StudentTeams = lazy(() => import('@/pages/student/Teams'))
const StudentContestRecords = lazy(() => import('@/pages/student/ContestRecords'))

// 通用页面
const Discussion = lazy(() => import('@/pages/common/Discussion'))
const Profile = lazy(() => import('@/pages/common/Profile'))
const NotFound = lazy(() => import('@/pages/common/NotFound'))

// CTF竞赛页面
const ContestDetail = lazy(() => import('@/pages/contest/ContestDetail'))
const ContestJeopardy = lazy(() => import('@/pages/contest/Jeopardy'))
const ContestAgentBattle = lazy(() => import('@/pages/contest/AgentBattle'))
const ContestSpectate = lazy(() => import('@/pages/contest/Spectate'))
const ContestReplay = lazy(() => import('@/pages/contest/Replay'))

/**
 * 根据角色重定向到默认页面
 */
function RoleBasedRedirect() {
  const { user, isLoggedIn } = useUserStore()
  
  if (!isLoggedIn || !user) {
    return <Navigate to="/login" replace />
  }
  
  const defaultRoute = RoleDefaultRoute[user.role]
  return <Navigate to={defaultRoute} replace />
}

/**
 * 应用路由配置
 */
export default function AppRouter() {
  return (
    <Routes>
      {/* 登录页 */}
      <Route path="/login" element={<Login />} />
      
      {/* 根路径重定向 */}
      <Route path="/" element={<RoleBasedRedirect />} />
      
      {/* 需要认证的路由 */}
      <Route element={<AuthGuard />}>
        <Route element={<MainLayout />}>
          
          {/* ====== 平台管理员路由 ====== */}
          <Route path="/admin">
            <Route path="schools" element={<AdminSchools />} />
            <Route path="images" element={<AdminImages />} />
            <Route path="configs" element={<AdminConfigs />} />
            <Route path="statistics" element={<AdminStatistics />} />
            <Route path="monitor" element={<AdminMonitor />} />
            <Route path="contests" element={<AdminContests />} />
            <Route path="challenges" element={<AdminChallenges />} />
            <Route path="reviews" element={<AdminReviews />} />
            <Route path="vulnerabilities" element={<AdminVulnerabilities />} />
          </Route>
          
          {/* ====== 学校管理员路由 ====== */}
          <Route path="/school">
            <Route path="teachers" element={<SchoolTeachers />} />
            <Route path="students" element={<SchoolStudents />} />
            <Route path="classes" element={<SchoolClasses />} />
            <Route path="statistics" element={<SchoolStatistics />} />
            <Route path="settings" element={<SchoolSettings />} />
            <Route path="contests" element={<SchoolContests />} />
            <Route path="challenges" element={<SchoolChallenges />} />
            <Route path="cross-school" element={<SchoolCrossSchoolContests />} />
          </Route>
          
          {/* ====== 教师路由 ====== */}
          <Route path="/teacher">
            <Route path="courses" element={<TeacherCourses />} />
            <Route path="courses/create" element={<TeacherCourseCreate />} />
            <Route path="courses/:id" element={<TeacherCourseDetail />} />
            <Route path="experiments" element={<TeacherExperiments />} />
            <Route path="experiments/create" element={<TeacherExperimentCreate />} />
            <Route path="experiments/:id/edit" element={<TeacherExperimentEdit />} />
            <Route path="grading" element={<TeacherGrading />} />
            <Route path="discussions" element={<TeacherDiscussionManage />} />
            <Route path="statistics" element={<TeacherStatistics />} />
            <Route path="contests" element={<TeacherContests />} />
            <Route path="contests/monitor" element={<TeacherContestMonitor />} />
            <Route path="challenges" element={<TeacherChallenges />} />
          </Route>
          
          {/* ====== 学生路由 ====== */}
          <Route path="/student">
            <Route path="courses" element={<StudentCourses />} />
            <Route path="courses/:id" element={<StudentCourseDetail />} />
            <Route path="experiments" element={<StudentExperiments />} />
            <Route path="experiments/:id" element={<StudentExperimentDetail />} />
            <Route path="grades" element={<StudentGrades />} />
            <Route path="contests" element={<StudentContests />} />
            <Route path="teams" element={<StudentTeams />} />
            <Route path="records" element={<StudentContestRecords />} />
          </Route>
          
          {/* ====== 实验环境（独立页面）====== */}
          <Route path="/experiment/:id/env" element={<StudentExperimentEnv />} />
          
          {/* ====== 竞赛页面 ====== */}
          <Route path="/contest">
            <Route path=":id" element={<ContestDetail />} />
            <Route path=":id/jeopardy" element={<ContestJeopardy />} />
            <Route path=":id/battle" element={<ContestAgentBattle />} />
            <Route path=":id/spectate" element={<ContestSpectate />} />
            <Route path=":id/replay" element={<ContestReplay />} />
          </Route>
          
          {/* ====== 通用页面 ====== */}
          <Route path="/discussion/:courseId" element={<Discussion />} />
          <Route path="/profile" element={<Profile />} />
          
        </Route>
      </Route>
      
      {/* 404 页面 */}
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
