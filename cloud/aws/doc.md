## 概要

AWS はあまり触ったことなかったので EKS を触ってみる

## 参考




## セットアップ

IAMでアクセスキーを生成しておく (一旦、IAM Identity Center を利用しない場合で試す)

aws cli
```
$ brew install awscli
$ aws configure --profile admin

```

クラスタ構築を試す
https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/getting-started-console.html

```
aws cloudformation create-stack \
  --region ap-northeast-1 \
  --stack-name my-eks-vpc-stack \
  --template-url https://s3.us-west-2.amazonaws.com/amazon-eks/cloudformation/2020-10-29/amazon-eks-vpc-private-subnets.yaml \
  --profile admin
```

```
$ aws iam create-role \
  --role-name myAmazonEKSClusterRole \
  --assume-role-policy-document file://"eks-cluster-role-trust-policy.json" \
  --profile admin

$ aws iam attach-role-policy \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSClusterPolicy \
  --role-name myAmazonEKSClusterRole \
  --profile admin

$ aws eks update-kubeconfig --region ap-northeast-1 --name my-cluster --profile admin --kubeconfig eks-kubeconfig
$ kubectl get svc --kubeconfig eks-kubeconfig
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.100.0.1   <none>        443/TCP   37m
```

この時点では Node がなかった。
```
$ kubectl --kubeconfig eks-kubeconfig get nodes
No resources found
```

nodegroup を作る
```
$ aws iam create-role \
  --role-name myAmazonEKSNodeRole \
  --assume-role-policy-document file://"node-role-trust-policy.json" \
  --profile admin

$ aws iam attach-role-policy \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy \
  --role-name myAmazonEKSNodeRole \
  --profile admin
$ aws iam attach-role-policy \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
  --role-name myAmazonEKSNodeRole \
  --profile admin
$ aws iam attach-role-policy \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy \
  --role-name myAmazonEKSNodeRole \
  --profile admin
```


## プライベートクラスタの作成

参考
https://eksctl.io/usage/eks-private-cluster/
https://github.com/NoppyOrg/EKSPoC_For_SecureEnvironment
https://www.alpha.co.jp/blog/202010_01/
https://milestone-of-se.nesuke.com/sv-advanced/aws/internet-nat-gateway/