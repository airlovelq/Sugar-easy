# Sugar-easy

#### 介绍
基于kubernetes crd（kubebuilder）封装的机器学习模型分布式训练和预测服务部署框架，使机器学习模型训练和服务部署流程标准化

#### 软件架构

主要思路为开发kubernetes crd controller，并定义自定义的训练和预测服务docker镜像，来控制训练和部署流程，使整个过程容器化。
整个工程按文件夹阐述，可分为以下几个部分：
1. image_define包含所需的docker镜像定义，包括训练镜像和预测服务镜像
2. kubernetes包含crd定义及其代码，配置文件
3. model_repository作为模型仓库，包含已封装的部分模型

#### 安装教程

1.  进入image_define文件夹，运行bash scripts/build.sh，打包自定义镜像
2.  需安装kubebuilder，https://github.com/kubernetes-sigs/kubebuilder/releases，链接中下载2.3.1版本
3.  进入kubernetes/appserver文件夹，make install, make deploy
4.  进入kubernetes/distributetrain文件夹，make install, make deploy
5.  如果需要使用预测服务的弹性伸缩，类似kubernetes Horizontal Pod Autoscaler（HPA），需要部署metrics-server，参照git clone https://github.com/kubernetes-incubator/metrics-server，也可以通过kube-prometheus监控更多指标，参照https://github.com/prometheus-operator/kube-prometheus.git

#### 使用说明

1.  预测服务样例见kubernetes/appserver/config/samples/sugar_v1_appserver.yaml，以下是各字段说明：
    replicasmin                 #最小副本数
    replicasmax                 #最大副本数
    modelfile                   #模型文件名
    modelparam                  #模型加载的参数，文件夹或文件（checkpoint）
    modelclass                  #模型文件中的模型类名
    port                        #端口配置（参照kubernetes service端口配置）
    modelvolume                 #模型文件所在volume（参照kubernetes volume配置）
    modelparamvolume            #模型加载的参数所在volume（参照kubernetes volume配置）
    resources                   #每个pod的资源配置（参照kubernetes resources配置）
    metrics                     #kubernetes HPA弹性伸缩监控指标配置（参照kubernetes metrics配置）
2.  训练样例见kubernetes/distributetrain/config/samples/sugar_v1_distributetrainjob.yaml，以下是各字段说明：
    replicas                    #副本数
    selector                    #参照kubernetes selector
    ports                       #分布式训练过程中使用的通信端口
    modelvolume                 #模型文件所在volume（参照kubernetes volume配置）
    modelparamvolume            #模型加载的参数所在volume（参照kubernetes volume配置）
    datasetvolume               #数据集所在volume（参照kubernetes volume配置）
    logvolume                   #日志所在volume（参照kubernetes volume配置）
    modelsavepath               #模型训练完成参数的保存路径，文件名或文件夹
    modelfile                   #模型文件名
    modelclass                  #模型文件中的模型类名
    modelcheckpoint             #模型的预训练参数  
    modelparams                 #模型的初始化参数
    trainparams                 #模型的可变训练参数设置，对于不启用自动调参的情况，每个value为单个值，对于启动自动调参的情况，每个value为列表或元祖
    traindataset                #训练数据集
    validatedataset             #验证数据集
    destscore                   #目标得分
    maxtrials                   #最大尝试次数，对于已启用自动调参的任务需要设置
    useautoml                   #是否使用自动调参
3.  模型开发说明，每个模型类需实现以下类函数：
    train(self, dataset_path, **train_args)   训练函数
    evaluate(self, dataset_path)              评估函数
    predict(self, query)                      预测函数
    save(self, path)                          保存训练（参数）结果
    load(self, path)                          加载模型训练（参数）结果
    destroy(self)                             模型结束运行时某些必要的资源释放


#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request


#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
