import{b6 as n,b as c}from"./index.9265efac.js";const i=(s="/",t="")=>n.post("/fs/get",{path:s,password:t}),d=(s="/",t="",e=1,o=0,a=!1,r)=>n.post("/fs/list",{path:s,password:t,page:e,per_page:o,refresh:a},{cancelToken:r}),u=(s="/",t="",e=!1)=>n.post("/fs/dirs",{path:s,password:t,force_root:e}),m=s=>n.post("/fs/mkdir",{path:s}),l=(s,t)=>n.post("/fs/rename",{path:s,name:t}),y=(s,t,e)=>n.post("/fs/move",{src_dir:s,dst_dir:t,names:e}),v=(s,t)=>n.post("/fs/recursive_move",{src_dir:s,dst_dir:t}),h=(s,t,e)=>n.post("/fs/copy",{src_dir:s,dst_dir:t,names:e}),x=(s,t)=>n.post("/fs/remove",{dir:s,names:t}),b=(s,t)=>n.put("/fs/put",void 0,{headers:{"File-Path":encodeURIComponent(s),Password:t}}),w=(s,t,e)=>n.post(`/fs/add_${e}`,{path:s,urls:t}),f=async(s,t=!0)=>{try{const e=await c.get(s,{responseType:"blob",params:t?{ts:new Date().getTime()}:void 0}),o=await e.data.text(),a=e.headers["content-type"];return{content:o,contentType:a}}catch(e){return t?await f(s,!1):{content:`Failed to fetch ${s}: ${e}`,contentType:""}}},T=async(s,t,e="",o=1,a=100)=>n.post("/fs/search",{parent:s,keywords:t,page:o,per_page:a,password:e}),g=async(s=["/"],t=-1)=>n.post("/admin/index/build",{paths:s,max_depth:t}),R=async(s=[],t=-1)=>n.post("/admin/index/update",{paths:s,max_depth:t});export{d as a,f as b,h as c,y as d,x as e,i as f,l as g,b as h,m as i,v as j,u as k,T as l,g as m,w as o,R as u};
