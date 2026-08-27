resource "aws_lb" "k8s_api" {
  name               = "${var.cluster_name}-nlb"
  internal           = false
  load_balancer_type = "network"

  subnets = aws_subnet.public[*].id

  security_groups = [
    aws_security_group.nlb.id
  ]

  enable_cross_zone_load_balancing = true

  tags = {
    Name = "${var.cluster_name}-nlb"
  }
}

resource "aws_lb_target_group" "k8s_api" {
  name        = "${var.cluster_name}-api"
  port        = 6443
  protocol    = "TCP"
  vpc_id      = aws_vpc.k8s.id
  target_type = "instance"

  health_check {
    protocol            = "TCP"
    port                = "6443"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 10
  }

  tags = {
    Name = "${var.cluster_name}-api-tg"
  }
}

resource "aws_lb_target_group_attachment" "masters" {
  count = 3

  target_group_arn = aws_lb_target_group.k8s_api.arn
  target_id        = aws_instance.master[count.index].id
  port             = 6443
}

resource "aws_lb_listener" "k8s_api" {
  load_balancer_arn = aws_lb.k8s_api.arn
  port              = 6443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.k8s_api.arn
  }
}
