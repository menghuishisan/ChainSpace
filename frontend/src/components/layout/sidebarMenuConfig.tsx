import type { ReactNode } from 'react'
import {
  AppstoreOutlined,
  AuditOutlined,
  BarChartOutlined,
  BookOutlined,
  BugOutlined,
  CommentOutlined,
  DatabaseOutlined,
  ExperimentOutlined,
  FileTextOutlined,
  SettingOutlined,
  SolutionOutlined,
  TeamOutlined,
  TrophyOutlined,
  UserOutlined,
} from '@ant-design/icons'

import type { SidebarMenuItem } from '@/types/presentation'

function createMenuItem(
  key: string,
  label: ReactNode,
  icon?: ReactNode,
  children?: SidebarMenuItem[],
): SidebarMenuItem {
  return { key, icon, children, label } as SidebarMenuItem
}

export function getSidebarMenuItems(role: string): SidebarMenuItem[] {
  switch (role) {
    case 'platform_admin':
      return [
        createMenuItem('/admin/schools', '学校管理', <TeamOutlined />),
        createMenuItem('/admin/images', '镜像管理', <AppstoreOutlined />),
        createMenuItem('/admin/configs', '系统配置', <SettingOutlined />),
        createMenuItem('/admin/statistics', '运营统计', <BarChartOutlined />),
        createMenuItem('/admin/monitor', '系统监控', <DatabaseOutlined />),
        createMenuItem('admin-contest', '竞赛管理', <TrophyOutlined />, [
          createMenuItem('/admin/contests', '平台比赛'),
          createMenuItem('/admin/challenges', '公共题库'),
        ]),
        createMenuItem('admin-review', '审核管理', <AuditOutlined />, [
          createMenuItem('/admin/reviews', '跨校比赛审核'),
        ]),
        createMenuItem('/admin/vulnerabilities', '漏洞转化', <BugOutlined />),
      ]
    case 'school_admin':
      return [
        createMenuItem('/school/teachers', '教师管理', <UserOutlined />),
        createMenuItem('/school/students', '学生管理', <SolutionOutlined />),
        createMenuItem('/school/classes', '班级管理', <TeamOutlined />),
        createMenuItem('/school/statistics', '本校统计', <BarChartOutlined />),
        createMenuItem('/school/settings', '学校设置', <SettingOutlined />),
        createMenuItem('school-contest', '竞赛管理', <TrophyOutlined />, [
          createMenuItem('/school/contests', '校内比赛'),
          createMenuItem('/school/challenges', '本校题库'),
          createMenuItem('/school/cross-school', '跨校比赛申请'),
        ]),
      ]
    case 'teacher':
      return [
        createMenuItem('teacher-course', '课程管理', <BookOutlined />, [
          createMenuItem('/teacher/courses', '我的课程'),
          createMenuItem('/teacher/courses/create', '创建课程'),
        ]),
        createMenuItem('teacher-experiment', '实验管理', <ExperimentOutlined />, [
          createMenuItem('/teacher/experiments', '实验列表'),
          createMenuItem('/teacher/grading', '批改作业'),
        ]),
        createMenuItem('/teacher/discussions', '讨论管理', <CommentOutlined />),
        createMenuItem('/teacher/statistics', '教学统计', <BarChartOutlined />),
        createMenuItem('teacher-contest', '竞赛中心', <TrophyOutlined />, [
          createMenuItem('/teacher/challenges', '题目管理'),
          createMenuItem('/teacher/contests', '比赛管理'),
          createMenuItem('/teacher/contests/monitor', '比赛监控'),
        ]),
      ]
    case 'student':
      return [
        createMenuItem('/student/courses', '我的课程', <BookOutlined />),
        createMenuItem('/student/experiments', '我的实验', <ExperimentOutlined />),
        createMenuItem('/student/grades', '我的成绩', <FileTextOutlined />),
        createMenuItem('student-contest', '竞赛中心', <TrophyOutlined />, [
          createMenuItem('/student/contests', '比赛列表'),
          createMenuItem('/student/teams', '我的队伍'),
          createMenuItem('/student/records', '竞赛记录'),
        ]),
      ]
    default:
      return []
  }
}

export function getSidebarSelection(
  pathname: string,
  items: SidebarMenuItem[],
): { selectedKeys: string[]; openKeys: string[] } {
  const selectedKeys: string[] = []
  const openKeys: string[] = []

  const findKeys = (menuItems: SidebarMenuItem[], parentKey?: string): boolean => {
    for (const item of menuItems) {
      if (!item) {
        continue
      }

      const menuItem = item as { key?: string; children?: SidebarMenuItem[] }
      if (menuItem.key === pathname) {
        selectedKeys.push(menuItem.key)
        if (parentKey) {
          openKeys.push(parentKey)
        }
        return true
      }

      if (menuItem.children && findKeys(menuItem.children, menuItem.key)) {
        return true
      }
    }

    return false
  }

  const findByPrefix = (menuItems: SidebarMenuItem[], parentKey?: string): boolean => {
    for (const item of menuItems) {
      if (!item) {
        continue
      }

      const menuItem = item as { key?: string; children?: SidebarMenuItem[] }
      if (typeof menuItem.key === 'string' && pathname.startsWith(menuItem.key)) {
        selectedKeys.push(menuItem.key)
        if (parentKey) {
          openKeys.push(parentKey)
        }
        return true
      }

      if (menuItem.children && findByPrefix(menuItem.children, menuItem.key)) {
        return true
      }
    }

    return false
  }

  findKeys(items)
  if (selectedKeys.length === 0) {
    findByPrefix(items)
  }

  return { selectedKeys, openKeys }
}
