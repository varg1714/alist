import{f as t,a0 as l,z as $,aE as h,aH as b,cm as u,m as f,aF as w,bv as j,ah as k,a7 as p,ag as y,bc as a,by as A,bI as v,d as C,e as d,K as S,cn as z,co as I,cp as O,ap as m,v as P,o as M,W as E}from"./index.c7be97ab.js";import{b as L}from"./Folder.36e2b55c.js";import{a as B,M as F}from"./Layout.8918bdbb.js";import{c as H}from"./useUtil.a37c2ef7.js";import{n as T}from"./index.abc04cae.js";import{g as W,O as D}from"./icon.988ad150.js";import"./Paginator.799ada6d.js";import"./index.f3072a2b.js";import"./api.2e4c6f66.js";import"./Markdown.b592c64e.js";import"./index.e5b7e6f6.js";import"./FolderTree.db0f2d89.js";const n=[{name:"name",textAlign:"left",w:{"@initial":"76%","@md":"50%"}},{name:"size",textAlign:"right",w:{"@initial":"24%","@md":"17%"}},{name:"modified",textAlign:"right",w:{"@initial":0,"@md":"33%"}}],_=e=>{const{isHide:o}=H();if(o(e.obj))return null;const{setPathAs:c}=B(),{show:s}=L({id:1});return t(F.div,{initial:{opacity:0,scale:.95},animate:{opacity:1,scale:1},transition:{duration:.2},style:{width:"100%"},get children(){return t(l,{class:"list-item",w:"$full",p:"$2",rounded:"$lg",transition:"all 0.3s",get _hover(){return{transform:"scale(1.01)",bgColor:$()}},as:T,get href(){return e.obj.name},onMouseEnter:()=>{c(e.obj.name,e.obj.is_dir,!0)},onContextMenu:r=>{h(()=>{b(!1),u(e.index,!0,!0)}),s(r,{props:e.obj})},get children(){return[t(l,{class:"name-box",spacing:"$1",get w(){return n[0].w},get children(){return[t(f,{get when(){return w()},get children(){return t(j,{"on:click":r=>{r.stopPropagation()},get checked(){return e.obj.selected},onChange:r=>{u(e.index,r.target.checked)}})}}),t(k,{class:"icon",boxSize:"$6",get color(){return p()},get as(){return W(e.obj)},mr:"$1","on:click":r=>{e.obj.type===D.IMAGE&&(r.stopPropagation(),r.preventDefault(),y.emit("gallery",e.obj.name))}}),t(a,{class:"name",css:{whiteSpace:"nowrap",overflow:"hidden",textOverflow:"ellipsis"},get title(){return e.obj.name},get children(){return e.obj.name}})]}}),t(a,{class:"size",get w(){return n[1].w},get textAlign(){return n[1].textAlign},get children(){return A(e.obj.size)}}),t(a,{class:"modified",display:{"@initial":"none","@md":"inline"},get w(){return n[2].w},get textAlign(){return n[2].textAlign},get children(){return v(e.obj.modified)}})]}})}})},ee=()=>{const e=C(),[o,c]=d(),[s,r]=d(!1);S(()=>{o()&&z(o(),s())});const g=i=>({fontWeight:"bold",fontSize:"$sm",color:"$neutral11",textAlign:i.textAlign,cursor:"pointer",onClick:()=>{i.name===o()?r(!s()):h(()=>{c(i.name),r(!1)})}});return t(E,{class:"list",w:"$full",spacing:"$1",get children(){return[t(l,{class:"title",w:"$full",p:"$2",get children(){return[t(l,{get w(){return n[0].w},spacing:"$1",get children(){return[t(f,{get when(){return w()},get children(){return t(j,{get checked(){return I()},get indeterminate(){return O()},onChange:i=>{b(i.target.checked)}})}}),t(a,m(()=>g(n[0]),{get children(){return e(`home.obj.${n[0].name}`)}}))]}}),t(a,m({get w(){return n[1].w}},()=>g(n[1]),{get children(){return e(`home.obj.${n[1].name}`)}})),t(a,m({get w(){return n[2].w}},()=>g(n[2]),{display:{"@initial":"none","@md":"inline"},get children(){return e(`home.obj.${n[2].name}`)}}))]}}),t(P,{get each(){return M.objs},children:(i,x)=>t(_,{obj:i,get index(){return x()}})})]}})};export{ee as default};
