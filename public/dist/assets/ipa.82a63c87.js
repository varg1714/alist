import{d as c,e as r,f as e,a0 as p,B as n,b9 as u,cv as g,o as a,cu as f}from"./index.3efbc392.js";import{a as d}from"./useUtil.d7d390bc.js";import{F as h}from"./File.e9ddce23.js";import"./api.0bcae678.js";import"./icon.a42451f6.js";import"./index.7cfe7831.js";import"./index.b028a38a.js";import"./Layout.4b8e34aa.js";import"./Markdown.d2f3b361.js";import"./index.80d83bf0.js";import"./FolderTree.ddbb4ab0.js";const U=()=>{const t=c(),[o,i]=r(!1),[s,l]=r(!1),{currentObjLink:m}=d();return e(h,{get children(){return e(p,{spacing:"$2",get children(){return[e(n,{as:"a",get href(){return`itms-services://?action=download-manifest&url=${u}/i/${g(encodeURIComponent(a.raw_url)+"/"+f(encodeURIComponent(a.obj.name)))}.plist`},onClick:()=>{i(!0)},get children(){return t(`home.preview.${o()?"installing":"install"}`)}}),e(n,{as:"a",colorScheme:"primary",get href(){return"apple-magnifier://install?url="+encodeURIComponent(m(!0))},onClick:()=>{l(!0)},get children(){return t(`home.preview.${s()?"tr-installing":"tr-install"}`)}})]}})}})};export{U as default};
