#coding=utf-8  
#上面是因为worker计算内容各不相同，不过在深度学习中，一般每个worker的计算内容都是一样的，  
# 都是计算神经网络的每个batch前向传导，所以一般代码是重用的  
import  tensorflow as tf  
#现在假设我们有A、B台机器，首先需要在各台机器上写一份代码，并跑起来，各机器上的代码内容大部分相同  
# 除了开始定义的时候，需要各自指定该台机器的task之外。以机器A为例子，A机器上的代码如下： 
import os
import time
import signal

# def myHandler(signum, frame):
#     try:
#         exit()
#     except BaseException:
#         pass

distribute_ips = os.environ.get('DistributeTrainJobIPs').split(',')
distribute_ports = os.environ.get('DistributeTrainJobPorts').split(',')
distribute_id = os.environ.get('DistributeTrainJobID')

ip_list = [ip+':'+distribute_ports[0] for ip in distribute_ips]
print(ip_list)
workers = ip_list[0:-1]

cluster_size = len(distribute_ips)

cluster=tf.train.ClusterSpec({  
    "worker": workers,
    # [  
    #     "192.168.11.105:1234",#格式 IP地址：端口号，第一台机器A的IP地址 ,在代码中需要用这台机器计算的时候，就要定义：/job:worker/task:0  
    # ],  
    "ps": [  
        ip_list[len(ip_list)-1]#第四台机器的IP地址 对应到代码块：/job:ps/task:0  
    ]})  
    
#不同的机器，下面这一行代码各不相同，server可以根据job_name、task_index两个参数，查找到集群cluster中对应的机器  
    
isps=(int(distribute_id)==(cluster_size-1))  
if isps:  
    server=tf.train.Server(cluster,job_name='ps',task_index=0)#找到‘ps’名字下的，task0，也就是机器A 
    # signal.signal(signal.SIGALRM, myHandler) 
    # signal.alarm(70)
    time.sleep(35)
    exit(0)
    # server.join()
else:  
    server=tf.train.Server(cluster,job_name='worker',task_index=int(distribute_id))#找到‘worker’名字下的，task0，也就是机器A  
    with tf.device(tf.train.replica_device_setter(worker_device='/job:worker/task:{}'.format(distribute_id),cluster=cluster)):  
        w=tf.get_variable('w',(2,2),tf.float32,initializer=tf.constant_initializer(2))  
        b=tf.get_variable('b',(2,2),tf.float32,initializer=tf.constant_initializer(5))  
        addwb=w+b  
        mutwb=w*b  
        divwb=w/b  
    
saver = tf.train.Saver()  
summary_op = tf.summary.merge_all() 
init_op = tf.initialize_all_variables()  
sv = tf.train.Supervisor(init_op=init_op, summary_op=summary_op, saver=saver)  
with sv.managed_session(server.target) as sess:  
    index = 0
    while index<20:  
        print(sess.run([addwb,mutwb,divwb])) 
        index += 1 
        time.sleep(1)
sv.stop()
