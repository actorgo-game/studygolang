export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  per_page: number
}

export interface User {
  uid: number
  username: string
  email: string
  name: string
  avatar: string
  city: string
  company: string
  github: string
  weibo: string
  website: string
  monlog: string
  introduce: string
  balance: number
  status: number
  is_root: boolean
  ctime: string
  mtime: string
  role: number
  role_name: string
}

export interface Me {
  uid: number
  username: string
  name: string
  avatar: string
  email: string
  status: number
  is_root: boolean
  balance: number
  msgnum: number
  is_vip: boolean
  role: number
}

export interface Topic {
  tid: number
  title: string
  content: string
  nid: number
  uid: number
  lastreplyuid: number
  lastreplytime: string
  editor_uid: number
  top: number
  tags: string
  flag: number
  ctime: string
  mtime: string
  cmtnum: number
  likenum: number
  viewnum: number
  closenum: number
  user: User
  node: TopicNode
  replies: Comment[]
}

export interface TopicNode {
  nid: number
  parent: number
  name: string
  intro: string
  seq: number
  ctime: string
}

export interface Article {
  id: number
  domain: string
  name: string
  title: string
  cover: string
  author: string
  author_txt: string
  lang: number
  pub_date: string
  url: string
  content: string
  txt: string
  tags: string
  viewnum: number
  cmtnum: number
  likenum: number
  status: number
  op_user: number
  ctime: string
  mtime: string
  user: User
  gctt_translated_at: string
  gctt_translator: User
}

export interface Resource {
  id: number
  title: string
  form: string
  content: string
  url: string
  uid: number
  catid: number
  ctime: string
  mtime: string
  cmtnum: number
  likenum: number
  viewnum: number
  status: number
  user: User
}

export interface Project {
  id: number
  name: string
  category: string
  uri: string
  home: string
  doc: string
  download: string
  src: string
  logo: string
  desc: string
  repo: string
  author: string
  licence: string
  lang: string
  os: string
  tags: string
  username: string
  viewnum: number
  cmtnum: number
  likenum: number
  status: number
  ctime: string
  user: User
}

export interface Book {
  id: number
  name: string
  ename: string
  cover: string
  author: string
  translator: string
  lang: number
  pub_date: string
  desc: string
  tags: string
  catalogue: string
  is_free: boolean
  online_url: string
  download_url: string
  buy_url: string
  price: number
  viewnum: number
  cmtnum: number
  likenum: number
  ctime: string
}

export interface Wiki {
  id: number
  title: string
  content: string
  uri: string
  uid: number
  tags: string
  viewnum: number
  cmtnum: number
  likenum: number
  ctime: string
  mtime: string
  user: User
}

export interface Reading {
  id: number
  rtype: number
  content: string
  inner: number
  url: string
  moreurls: string
  username: string
  clicknum: number
  ctime: string
  urls: string[]
}

export interface Comment {
  cid: number
  objid: number
  objtype: number
  content: string
  uid: number
  floor: number
  likenum: number
  ctime: string
  user: User
}

export interface Message {
  id: number
  from: number
  to: number
  content: string
  hasread: string
  ctime: string
  from_user: User
  to_user: User
}

export interface SiteStat {
  user: number
  topic: number
  article: number
  resource: number
  project: number
  book: number
  comment: number
}

export interface FriendLink {
  id: number
  name: string
  url: string
  logo: string
  seq: number
}

export interface Mission {
  id: number
  name: string
  award: number
  state: number
}

export interface Gift {
  id: number
  name: string
  desc: string
  price: number
  image: string
  total_num: number
  remain_num: number
}
