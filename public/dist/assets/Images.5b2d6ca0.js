import{f as t,a7 as n,ah as m,e as b,t as h,aF as f,W as j,aE as x,aH as $,cm as s,ak as k,m as p,bv as w,af as v,ag as y,v as C,o as I,bg as M}from"./index.86895a3d.js";import{b as E}from"./Folder.7c5833e9.js";import{a as L,M as S}from"./Layout.c6971c82.js";import{c as F,a as O}from"./useUtil.82a9a0ca.js";import{I as z}from"./ImageWithError.7e0e79de.js";import{O as A,g as H}from"./icon.7c2d87d9.js";import"./Paginator.297c36de.js";import"./index.bd520d27.js";import"./index.29bdb5b4.js";import"./api.0999634a.js";import"./Markdown.d3253182.js";import"./index.4c2446cc.js";import"./FolderTree.844c3ed2.js";const W=e=>{const{isHide:r}=F();if(r(e.obj)||e.obj.type!==A.IMAGE)return null;const{setPathAs:i}=L(),l=t(m,{get color(){return n()},boxSize:"$12",get as(){return H(e.obj)}}),[c,o]=b(!1),u=h(()=>f()&&(c()||e.obj.selected)),{show:g}=E({id:1}),{rawLink:d}=O();return t(S.div,{initial:{opacity:0,scale:.9},animate:{opacity:1,scale:1},transition:{duration:.2},style:{"flex-grow":1},get children(){return t(j,{w:"$full",class:"image-item",p:"$1",spacing:"$1",rounded:"$lg",transition:"all 0.3s",border:"2px solid transparent",get _hover(){return{border:`2px solid ${n()}`}},onMouseEnter:()=>{o(!0),i(e.obj.name,e.obj.is_dir,!0)},onMouseLeave:()=>{o(!1)},onContextMenu:a=>{x(()=>{$(!1),s(e.index,!0,!0)}),g(a,{props:e.obj})},get children(){return t(k,{w:"$full",pos:"relative",get children(){return[t(p,{get when(){return u()},get children(){return t(w,{pos:"absolute",left:"$1",top:"$1",get checked(){return e.obj.selected},onChange:a=>{s(e.index,a.target.checked)}})}}),t(z,{h:"150px",w:"$full",objectFit:"cover",rounded:"$lg",shadow:"$md",get fallback(){return t(v,{size:"lg"})},fallbackErr:l,get src(){return d(e.obj)},loading:"lazy",onClick:()=>{y.emit("gallery",e.obj.name)}})]}})}})}})},R=()=>t(M,{w:"$full",gap:"$1",flexWrap:"wrap",get children(){return t(C,{get each(){return I.objs},children:(e,r)=>t(W,{obj:e,get index(){return r()}})})}});export{R as default};
